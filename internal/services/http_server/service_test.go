package http_server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"testing"

	"viktig/internal/entities"
	"viktig/internal/queue"
	"viktig/internal/repository"
	"viktig/internal/repository/in_memory_repo"
	"viktig/internal/services/http_server/handlers"
	"viktig/internal/services/http_server/handlers/debug_handler"
	"viktig/internal/services/http_server/handlers/vk_callback_handler"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

type handlerStub struct{}

func (s handlerStub) Handle(_ *fasthttp.RequestCtx) {}

func TestServiceStartStop(t *testing.T) {
	h := &handlers.Handlers{Metrics: &handlerStub{}, VkCallbackHandler: &handlerStub{}}
	s := New("localhost", "1337", h, slog.Default())

	var ok bool
	p := gomonkey.ApplyFunc(net.Listen, func(network string, address string) (net.Listener, error) {
		if ok {
			return fasthttputil.NewInmemoryListener(), nil
		}
		return nil, fmt.Errorf("error")
	})
	defer p.Reset()

	ok = false
	err := s.Run(context.Background())
	assert.EqualError(t, err, "error")

	ok = true
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		errCh <- s.Run(ctx)
	}()
	cancel()
	assert.NoError(t, <-errCh)
}

func TestVkHandler(t *testing.T) {
	setupVkHandlerTest := func(t *testing.T, interactionId uuid.UUID, confirmationString string) (*queue.Queue[entities.Message], *http.Client) {
		t.Helper()
		listener := fasthttputil.NewInmemoryListener()
		p := gomonkey.ApplyFunc(net.Listen, func(network string, address string) (net.Listener, error) { return listener, nil })
		t.Cleanup(p.Reset)

		q := queue.NewQueue[entities.Message]()
		repo := in_memory_repo.New()
		_ = repo.StoreUser(&entities.User{Id: 123})
		_ = repo.StoreInteraction(&entities.Interaction{
			Id:                 interactionId,
			UserId:             123,
			ConfirmationString: confirmationString,
			TgChatId:           1234,
		})
		h := &handlers.Handlers{Metrics: &handlerStub{}, VkCallbackHandler: vk_callback_handler.New(q, repo, slog.Default()), Debug: &handlerStub{}}

		s := New("localhost", "1337", h, slog.Default())

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		go func() { _ = s.Run(ctx) }()

		client := &http.Client{Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return listener.Dial()
			},
		}}
		return q, client
	}

	type args struct {
		interactionId      uuid.UUID
		confirmationString string
		reqBody            []byte
		url                string
	}
	interactionId, _ := uuid.NewRandom()
	interactionId2, _ := uuid.NewRandom()
	tests := []struct {
		name          string
		args          args
		statusCode    int
		bodyContains  []string
		queueContents []entities.Message
	}{
		// general:
		{
			name: "json unmarshal error",
			args: args{
				reqBody: []byte(`{"type":"",}`),
				url:     fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId.String()),
			},
			statusCode:   400,
			bodyContains: []string{"json unmarshal error"},
		},
		{
			name: "unsupported message type",
			args: args{
				reqBody: []byte(`{"type":"unsupported_message_type","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:     fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId.String()),
			},
			statusCode:   400,
			bodyContains: []string{"unsupported message type: unsupported_message_type"},
		},
		// handleChallenge:
		{
			name: "handleChallenge: no interactionId",
			args: args{
				reqBody: []byte(`{"type":"confirmation","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:     "http://localhost:8080/api/vk/callback",
			},
			statusCode:   404,
			bodyContains: []string{"Not Found"},
		},
		{
			name: "handleChallenge: bad interactionId",
			args: args{
				reqBody: []byte(`{"type":"confirmation","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:     "http://localhost:8080/api/vk/callback/invalid-uuid",
			},
			statusCode:   400,
			bodyContains: []string{"invalid interactionId"},
		},
		{
			name: "handleChallenge: interaction not found",
			args: args{
				interactionId: interactionId,
				reqBody:       []byte(`{"type":"confirmation","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:           fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId2.String()),
			},
			statusCode:   400,
			bodyContains: []string{"interaction not found"},
		},
		{
			name: "handleChallenge: ok",
			args: args{
				interactionId:      interactionId,
				confirmationString: "confirmationString",
				reqBody:            []byte(`{"type":"confirmation","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:                fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId.String()),
			},
			statusCode:   200,
			bodyContains: []string{"confirmationString"},
		},
		// handleMessage:
		{
			name: "handleMessage[message_new]: no interactionId",
			args: args{
				reqBody: []byte(`{"type":"message_new","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:     "http://localhost:8080/api/vk/callback",
			},
			statusCode:   404,
			bodyContains: []string{"Not Found"},
		},
		{
			name: "handleMessage[message_edit]: no interactionId",
			args: args{
				reqBody: []byte(`{"type":"message_edit","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:     "http://localhost:8080/api/vk/callback",
			},
			statusCode:   404,
			bodyContains: []string{"Not Found"},
		},
		{
			name: "handleMessage[message_reply]: no interactionId",
			args: args{
				reqBody: []byte(`{"type":"message_reply","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:     "http://localhost:8080/api/vk/callback",
			},
			statusCode:   404,
			bodyContains: []string{"Not Found"},
		},
		{
			name: "handleMessage[message_new]: json unmarshal error",
			args: args{
				reqBody: []byte(`{"type":"message_new","event":"test_event_id","v":"1.0","group_id":12345,"object":{"message":{"from_id":1,"text":"test text}}}`),
				url:     fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId.String()),
			},
			statusCode:   400,
			bodyContains: []string{"json unmarshal error"},
		},
		{
			name: "handleMessage[message_edit]: json unmarshal error",
			args: args{
				reqBody: []byte(`{"type":"message_edit","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1,"text":"test text}}`),
				url:     fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId.String()),
			},
			statusCode:   400,
			bodyContains: []string{"json unmarshal error"},
		},
		{
			name: "handleMessage[message_reply]: json unmarshal error",
			args: args{
				reqBody: []byte(`{"type":"message_reply","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1,"text":"test text}}`),
				url:     fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId.String()),
			},
			statusCode:   400,
			bodyContains: []string{"json unmarshal error"},
		},
		{
			name: "handleMessage[message_new]: bad message",
			args: args{
				reqBody: []byte(`{"type":"message_new","event":"test_event_id","v":"1.0","group_id":12345,"object":{"message":{"from_id":1}}}`),
				url:     fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId.String()),
			},
			statusCode:   400,
			bodyContains: []string{"validation error"},
		},
		{
			name: "handleMessage[message_edit]: json unmarshal error",
			args: args{
				reqBody: []byte(`{"type":"message_edit","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1}}`),
				url:     fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId.String()),
			},
			statusCode:   400,
			bodyContains: []string{"validation error"},
		},
		{
			name: "handleMessage[message_reply]: json unmarshal error",
			args: args{
				reqBody: []byte(`{"type":"message_reply","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1}}`),
				url:     fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId.String()),
			},
			statusCode:   400,
			bodyContains: []string{"validation error"},
		},
		{
			name: "handleMessage[message_new]: interaction does not exist",
			args: args{
				interactionId: interactionId,
				reqBody:       []byte(`{"type":"message_new","event":"test_event_id","v":"1.0","group_id":12345,"object":{"message":{"from_id":1,"text":"test text"}}}`),
				url:           fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId2.String()),
			},
			statusCode:   400,
			bodyContains: []string{"interaction does not exist"},
		},
		{
			name: "handleMessage[message_edit]: interaction does not exist",
			args: args{
				interactionId: interactionId,
				reqBody:       []byte(`{"type":"message_edit","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1,"text":"test text"}}`),
				url:           fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId2.String()),
			},
			statusCode:   400,
			bodyContains: []string{"interaction does not exist"},
		},
		{
			name: "handleMessage[message_reply]: interaction does not exist",
			args: args{
				interactionId: interactionId,
				reqBody:       []byte(`{"type":"message_reply","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1,"text":"test text"}}`),
				url:           fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId2.String()),
			},
			statusCode:   400,
			bodyContains: []string{"interaction does not exist"},
		},
		{
			name: "handleMessage[message_new]: ok",
			args: args{
				interactionId: interactionId,
				reqBody:       []byte(`{"type":"message_new","event":"test_event_id","v":"1.0","group_id":12345,"object":{"message":{"from_id":1,"text":"test text"}}}`),
				url:           fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId.String()),
			},
			statusCode:   200,
			bodyContains: []string{"ok"},
			queueContents: []entities.Message{{
				InteractionId: interactionId,
				Type:          entities.MessageTypeNew,
				Text:          "test text",
				VkSenderId:    1,
			}},
		},
		{
			name: "handleMessage[message_edit]: ok",
			args: args{
				interactionId: interactionId,
				reqBody:       []byte(`{"type":"message_edit","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1,"text":"test text"}}`),
				url:           fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId.String()),
			},
			statusCode:   200,
			bodyContains: []string{"ok"},
			queueContents: []entities.Message{{
				InteractionId: interactionId,
				Type:          entities.MessageTypeEdit,
				Text:          "test text",
				VkSenderId:    1,
			}},
		},
		{
			name: "handleMessage[message_reply]: ok",
			args: args{
				interactionId: interactionId,
				reqBody:       []byte(`{"type":"message_reply","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1,"text":"test text"}}`),
				url:           fmt.Sprintf("http://localhost:8080/api/vk/callback/%s", interactionId.String()),
			},
			statusCode:   200,
			bodyContains: []string{"ok"},
			queueContents: []entities.Message{{
				InteractionId: interactionId,
				Type:          entities.MessageTypeReply,
				Text:          "test text",
				VkSenderId:    1,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, client := setupVkHandlerTest(t, tt.args.interactionId, tt.args.confirmationString)
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				var values []entities.Message = nil
				for {
					msg, ok := <-q.AsChan()
					if !ok {
						assert.Equal(t, tt.queueContents, values)
						return
					}

					if values == nil {
						values = []entities.Message{msg}
						continue
					}
					values = append(values, msg)
				}
			}()

			resp, err := client.Post(tt.args.url, "application/json", bytes.NewReader(tt.args.reqBody))
			close(q.AsChan())
			assert.NoError(t, err)
			respBodyBytes, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			respBody := string(respBodyBytes)

			assert.Equal(t, tt.statusCode, resp.StatusCode)
			for _, contains := range tt.bodyContains {
				assert.Contains(t, respBody, contains)
			}
			wg.Wait()
		})
	}
}

func TestDebugHandler(t *testing.T) {
	setupDebugHandlerTest := func(t *testing.T, host string) (repository.Repository, *http.Client) {
		t.Helper()
		listener := fasthttputil.NewInmemoryListener()
		p := gomonkey.ApplyFunc(net.Listen, func(network string, address string) (net.Listener, error) { return listener, nil })
		t.Cleanup(p.Reset)

		repo := in_memory_repo.New()
		h := &handlers.Handlers{
			Metrics:           &handlerStub{},
			VkCallbackHandler: vk_callback_handler.New(queue.NewQueue[entities.Message](), repo, slog.Default()),
			Debug:             debug_handler.New(host, repo),
		}

		s := New("localhost", "1337", h, slog.Default())
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		go func() { _ = s.Run(ctx) }()

		client := &http.Client{Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return listener.Dial()
			},
		}}
		return repo, client
	}
	doRequest := func(client *http.Client, requestBody []byte) (int, string) {
		resp, err := client.Post("http://localhost:8080/debug", "application/json", bytes.NewReader(requestBody))
		assert.NoError(t, err)
		respBodyBytes, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		return resp.StatusCode, string(respBodyBytes)
	}

	t.Run("json unmarshal error", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"test action",}`))

		assert.Equal(t, 400, code)
		assert.Contains(t, respBody, "json unmarshal error")
	})
	t.Run("unknown action", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"test action"}`))

		assert.Equal(t, 200, code)
		assert.Equal(t, "", respBody)
	})

	// New User
	t.Run("new user: no request data", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"new_user"}`))

		assert.Equal(t, 400, code)
		assert.Equal(t, "request data is nil", respBody)
	})
	t.Run("new user: insufficient data", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"new_user","data":{}}`))

		assert.Equal(t, 400, code)
		assert.Contains(t, respBody, "validation error")
	})
	t.Run("new user: excess", func(t *testing.T) {
		repo, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"new_user","data":{"id":123,"confirmation_string":"test_confirmation_string","tg_chat_id":321}}`))

		user, err := repo.GetUser(123)
		assert.NoError(t, err)
		assert.Equal(t, &entities.User{Id: 123}, user)
		assert.Equal(t, 200, code)
		assert.Equal(t, "ok", respBody)
	})
	t.Run("new user: ok", func(t *testing.T) {
		repo, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"new_user","data":{"id":123}}`))

		user, err := repo.GetUser(123)
		assert.NoError(t, err)
		assert.Equal(t, &entities.User{Id: 123}, user)
		assert.Equal(t, 200, code)
		assert.Equal(t, "ok", respBody)
	})
	// Get User
	t.Run("get user: no request data", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"get_user"}`))

		assert.Equal(t, 400, code)
		assert.Equal(t, "request data is nil", respBody)
	})
	t.Run("get user: insufficient data", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"get_user","data":{}}`))

		assert.Equal(t, 400, code)
		assert.Contains(t, respBody, "validation error")
	})
	t.Run("get user: does not exist", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"get_user","data":{"id":123}}`))

		assert.Equal(t, 400, code)
		assert.Equal(t, "user not found", respBody)
	})
	t.Run("get user: excess", func(t *testing.T) {
		repo, client := setupDebugHandlerTest(t, "http://localhost")
		_ = repo.StoreUser(&entities.User{Id: 123})

		code, respBody := doRequest(client, []byte(`{"action":"get_user","data":{"id":123,"confirmation_string":"test_confirmation_string","tg_chat_id":321}}`))

		assert.Equal(t, 200, code)
		assert.Equal(t, `{"id":123}`, respBody)
	})
	t.Run("get user: ok", func(t *testing.T) {
		repo, client := setupDebugHandlerTest(t, "http://localhost")
		_ = repo.StoreUser(&entities.User{Id: 123})

		code, respBody := doRequest(client, []byte(`{"action":"get_user","data":{"id":123}}`))

		assert.Equal(t, 200, code)
		assert.Equal(t, `{"id":123}`, respBody)
	})
	// Delete User
	t.Run("delete user: no request data", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"delete_user"}`))

		assert.Equal(t, 400, code)
		assert.Equal(t, "request data is nil", respBody)
	})
	t.Run("delete user: insufficient data", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"delete_user","data":{}}`))

		assert.Equal(t, 400, code)
		assert.Contains(t, respBody, "validation error")
	})
	t.Run("delete user: does not exist", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"delete_user","data":{"id":123}}`))

		assert.Equal(t, 400, code)
		assert.Equal(t, "user not found", respBody)
	})
	t.Run("delete user: excess", func(t *testing.T) {
		repo, client := setupDebugHandlerTest(t, "http://localhost")
		_ = repo.StoreUser(&entities.User{Id: 123})

		code, respBody := doRequest(client, []byte(`{"action":"delete_user","data":{"id":123,"confirmation_string":"test_confirmation_string","tg_chat_id":321}}`))

		user, err := repo.GetUser(123)
		assert.Nil(t, user)
		assert.EqualError(t, err, "user not found")
		assert.Equal(t, 200, code)
		assert.Equal(t, `{"id":123}`, respBody)
	})
	t.Run("delete user: ok", func(t *testing.T) {
		repo, client := setupDebugHandlerTest(t, "http://localhost")
		_ = repo.StoreUser(&entities.User{Id: 123})

		code, respBody := doRequest(client, []byte(`{"action":"delete_user","data":{"id":123}}`))

		user, err := repo.GetUser(123)
		assert.Nil(t, user)
		assert.EqualError(t, err, "user not found")
		assert.Equal(t, 200, code)
		assert.Equal(t, `{"id":123}`, respBody)
	})
	// New Interaction
	t.Run("new interaction: no request data", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"new_interaction"}`))

		assert.Equal(t, 400, code)
		assert.Equal(t, "request data is nil", respBody)
	})
	t.Run("new interaction: insufficient data", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"new_interaction","data":{"confirmation_string":"test_confirmation_string","tg_chat_id":321}}}`))

		assert.Equal(t, 400, code)
		assert.Contains(t, respBody, "validation error")
	})
	t.Run("new interaction: no user", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"new_interaction","data":{"user_id":123,"name":"test interaction","confirmation_string":"test_confirmation_string","tg_chat_id":321}}`))

		assert.Equal(t, 400, code)
		assert.Equal(t, "user not found", respBody)
	})
	t.Run("new interaction: excess", func(t *testing.T) {
		interactionId := uuid.New()
		p := gomonkey.ApplyFunc(uuid.New, func() uuid.UUID { return interactionId })
		defer p.Reset()
		repo, client := setupDebugHandlerTest(t, "http://localhost")
		_ = repo.StoreUser(&entities.User{Id: 123})

		code, respBody := doRequest(client, []byte(`{"action":"new_interaction","data":{"extra":"field","user_id":123,"name":"test interaction","confirmation_string":"test_confirmation_string","tg_chat_id":321}}`))

		interaction, err := repo.GetInteraction(interactionId)
		assert.NoError(t, err)
		assert.Equal(t, "test_confirmation_string", interaction.ConfirmationString)
		assert.Equal(t, 200, code)
		assert.Equal(t, fmt.Sprintf(`{"callback_url":"http://localhost/api/vk/callback/%s"}`, interactionId), respBody)
	})
	t.Run("new interaction: ok", func(t *testing.T) {
		interactionId := uuid.New()
		p := gomonkey.ApplyFunc(uuid.New, func() uuid.UUID { return interactionId })
		defer p.Reset()
		repo, client := setupDebugHandlerTest(t, "http://localhost")
		_ = repo.StoreUser(&entities.User{Id: 123})

		code, respBody := doRequest(client, []byte(`{"action":"new_interaction","data":{"user_id":123,"name":"test interaction","confirmation_string":"test_confirmation_string","tg_chat_id":321}}`))

		interaction, err := repo.GetInteraction(interactionId)
		assert.NoError(t, err)
		assert.Equal(t, "test_confirmation_string", interaction.ConfirmationString)
		assert.Equal(t, 200, code)
		assert.Equal(t, fmt.Sprintf(`{"callback_url":"http://localhost/api/vk/callback/%s"}`, interactionId), respBody)
	})
	// Get Interaction
	t.Run("get interaction: no request data", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"get_interaction"}`))

		assert.Equal(t, 400, code)
		assert.Equal(t, "request data is nil", respBody)
	})
	t.Run("get interaction: insufficient data", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"get_interaction","data":{}}`))

		assert.Equal(t, 400, code)
		assert.Contains(t, respBody, "validation error")
	})
	t.Run("get interaction: does not exist", func(t *testing.T) {
		repo, client := setupDebugHandlerTest(t, "http://localhost")
		interactionId := uuid.New()
		_ = repo.StoreUser(&entities.User{Id: 123})
		_ = repo.StoreInteraction(&entities.Interaction{
			Id:                 interactionId,
			Name:               "test name",
			UserId:             123,
			ConfirmationString: "test confirmation string",
			TgChatId:           321,
		})

		code, respBody := doRequest(client, []byte(fmt.Sprintf(`{"action":"get_interaction","data":{"id":"%s"}}`, uuid.New())))

		assert.Equal(t, 400, code)
		assert.Equal(t, "interaction not found", respBody)
	})
	t.Run("get interaction: excess", func(t *testing.T) {
		repo, client := setupDebugHandlerTest(t, "http://localhost")
		interactionId := uuid.New()
		_ = repo.StoreUser(&entities.User{Id: 123})
		_ = repo.StoreInteraction(&entities.Interaction{
			Id:                 interactionId,
			Name:               "test name",
			UserId:             123,
			ConfirmationString: "test confirmation string",
			TgChatId:           321,
		})

		code, respBody := doRequest(client, []byte(fmt.Sprintf(`{"action":"get_interaction","data":{"id":"%s","excess":"field"}}`, interactionId)))

		assert.Equal(t, 200, code)
		assert.Equal(t, fmt.Sprintf(`{"id":"%s","name":"test name","user_id":123,"confirmation_string":"test confirmation string","tg_chat_id":321}`, interactionId), respBody)
	})
	t.Run("get interaction: ok", func(t *testing.T) {
		repo, client := setupDebugHandlerTest(t, "http://localhost")
		interactionId := uuid.New()
		_ = repo.StoreUser(&entities.User{Id: 123})
		_ = repo.StoreInteraction(&entities.Interaction{
			Id:                 interactionId,
			Name:               "test name",
			UserId:             123,
			ConfirmationString: "test confirmation string",
			TgChatId:           321,
		})

		code, respBody := doRequest(client, []byte(fmt.Sprintf(`{"action":"get_interaction","data":{"id":"%s"}}`, interactionId)))

		assert.Equal(t, 200, code)
		assert.Equal(t, fmt.Sprintf(`{"id":"%s","name":"test name","user_id":123,"confirmation_string":"test confirmation string","tg_chat_id":321}`, interactionId), respBody)
	})
	// Delete Interaction
	t.Run("delete interaction: no request data", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"delete_interaction"}`))

		assert.Equal(t, 400, code)
		assert.Equal(t, "request data is nil", respBody)
	})
	t.Run("delete interaction: insufficient data", func(t *testing.T) {
		_, client := setupDebugHandlerTest(t, "http://localhost")

		code, respBody := doRequest(client, []byte(`{"action":"delete_interaction","data":{}}`))

		assert.Equal(t, 400, code)
		assert.Contains(t, respBody, "validation error")
	})
	t.Run("delete interaction: does not exist", func(t *testing.T) {
		repo, client := setupDebugHandlerTest(t, "http://localhost")
		interactionId := uuid.New()
		_ = repo.StoreUser(&entities.User{Id: 123})
		_ = repo.StoreInteraction(&entities.Interaction{
			Id:                 interactionId,
			Name:               "test name",
			UserId:             123,
			ConfirmationString: "test confirmation string",
			TgChatId:           321,
		})

		code, respBody := doRequest(client, []byte(fmt.Sprintf(`{"action":"delete_interaction","data":{"id":"%s"}}`, uuid.New())))

		assert.Equal(t, 400, code)
		assert.Equal(t, "interaction not found", respBody)
	})
	t.Run("delete interaction: excess", func(t *testing.T) {
		repo, client := setupDebugHandlerTest(t, "http://localhost")
		interactionId := uuid.New()
		_ = repo.StoreUser(&entities.User{Id: 123})
		_ = repo.StoreInteraction(&entities.Interaction{
			Id:                 interactionId,
			Name:               "test name",
			UserId:             123,
			ConfirmationString: "test confirmation string",
			TgChatId:           321,
		})

		code, respBody := doRequest(client, []byte(fmt.Sprintf(`{"action":"delete_interaction","data":{"id":"%s"}}`, interactionId)))

		interaction, err := repo.GetInteraction(interactionId)
		assert.Nil(t, interaction)
		assert.EqualError(t, err, "interaction not found")
		assert.Equal(t, 200, code)
		assert.Equal(t, fmt.Sprintf(`{"id":"%s","name":"test name","user_id":123,"confirmation_string":"test confirmation string","tg_chat_id":321}`, interactionId), respBody)
	})
	t.Run("delete interaction: ok", func(t *testing.T) {
		repo, client := setupDebugHandlerTest(t, "http://localhost")
		interactionId := uuid.New()
		_ = repo.StoreUser(&entities.User{Id: 123})
		_ = repo.StoreInteraction(&entities.Interaction{
			Id:                 interactionId,
			Name:               "test name",
			UserId:             123,
			ConfirmationString: "test confirmation string",
			TgChatId:           321,
		})

		code, respBody := doRequest(client, []byte(fmt.Sprintf(`{"action":"delete_interaction","data":{"id":"%s"}}`, interactionId)))

		interaction, err := repo.GetInteraction(interactionId)
		assert.Nil(t, interaction)
		assert.EqualError(t, err, "interaction not found")
		assert.Equal(t, 200, code)
		assert.Equal(t, fmt.Sprintf(`{"id":"%s","name":"test name","user_id":123,"confirmation_string":"test confirmation string","tg_chat_id":321}`, interactionId), respBody)
	})
}
