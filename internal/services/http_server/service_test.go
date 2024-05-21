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
	"viktig/internal/services/http_server/handlers"
	"viktig/internal/services/http_server/handlers/vk_callback_handler"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

type handlerStub struct{}

func (s handlerStub) Handle(_ *fasthttp.RequestCtx) {}

func TestServiceStartStop(t *testing.T) {
	h := &handlers.Handlers{Metrics: &handlerStub{}, VkCallbackHandler: &handlerStub{}}
	s := New(&Config{Host: "localhost", Port: 1337}, h, slog.Default())

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
	type args struct {
		interactionId      string
		confirmationString string
		reqBody            []byte
		url                string
	}
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
				url:     "http://localhost:8080/api/vk/callback/interactionId",
			},
			statusCode:   400,
			bodyContains: []string{"json unmarshal error"},
		},
		{
			name: "unsupported message type",
			args: args{
				reqBody: []byte(`{"type":"unsupported_message_type","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:     "http://localhost:8080/api/vk/callback/interactionId",
			},
			statusCode:   400,
			bodyContains: []string{"unsupported message type: unsupported_message_type"},
		},
		// handleChallenge:
		{
			name: "handleChallenge: no hook id",
			args: args{
				reqBody: []byte(`{"type":"confirmation","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:     "http://localhost:8080/api/vk/callback",
			},
			statusCode:   404,
			bodyContains: []string{"Not Found"},
		},
		{
			name: "handleChallenge: interaction not found",
			args: args{
				interactionId: "interactionId",
				reqBody:       []byte(`{"type":"confirmation","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:           "http://localhost:8080/api/vk/callback/interactionId2",
			},
			statusCode:   400,
			bodyContains: []string{"interaction not found"},
		},
		{
			name: "handleChallenge: ok",
			args: args{
				interactionId:      "interactionId",
				confirmationString: "confirmationString",
				reqBody:            []byte(`{"type":"confirmation","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:                "http://localhost:8080/api/vk/callback/interactionId",
			},
			statusCode:   200,
			bodyContains: []string{"confirmationString"},
		},
		// handleMessage:
		{
			name: "handleMessage[message_new]: no hook id",
			args: args{
				reqBody: []byte(`{"type":"message_new","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:     "http://localhost:8080/api/vk/callback",
			},
			statusCode:   404,
			bodyContains: []string{"Not Found"},
		},
		{
			name: "handleMessage[message_edit]: no hook id",
			args: args{
				reqBody: []byte(`{"type":"message_edit","event":"test_event_id","v":"1.0","group_id":12345}`),
				url:     "http://localhost:8080/api/vk/callback",
			},
			statusCode:   404,
			bodyContains: []string{"Not Found"},
		},
		{
			name: "handleMessage[message_reply]: no hook id",
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
				url:     "http://localhost:8080/api/vk/callback/interactionId",
			},
			statusCode:   400,
			bodyContains: []string{"json unmarshal error"},
		},
		{
			name: "handleMessage[message_edit]: json unmarshal error",
			args: args{
				reqBody: []byte(`{"type":"message_edit","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1,"text":"test text}}`),
				url:     "http://localhost:8080/api/vk/callback/interactionId",
			},
			statusCode:   400,
			bodyContains: []string{"json unmarshal error"},
		},
		{
			name: "handleMessage[message_reply]: json unmarshal error",
			args: args{
				reqBody: []byte(`{"type":"message_reply","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1,"text":"test text}}`),
				url:     "http://localhost:8080/api/vk/callback/interactionId",
			},
			statusCode:   400,
			bodyContains: []string{"json unmarshal error"},
		},
		{
			name: "handleMessage[message_new]: interaction does not exist",
			args: args{
				interactionId: "interactionId",
				reqBody:       []byte(`{"type":"message_new","event":"test_event_id","v":"1.0","group_id":12345,"object":{"message":{"from_id":1,"text":"test text"}}}`),
				url:           "http://localhost:8080/api/vk/callback/interactionId2",
			},
			statusCode:   400,
			bodyContains: []string{"interaction does not exist"},
		},
		{
			name: "handleMessage[message_edit]: interaction does not exist",
			args: args{
				interactionId: "interactionId",
				reqBody:       []byte(`{"type":"message_edit","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1,"text":"test text"}}`),
				url:           "http://localhost:8080/api/vk/callback/interactionId2",
			},
			statusCode:   400,
			bodyContains: []string{"interaction does not exist"},
		},
		{
			name: "handleMessage[message_reply]: interaction does not exist",
			args: args{
				interactionId: "interactionId",
				reqBody:       []byte(`{"type":"message_reply","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1,"text":"test text"}}`),
				url:           "http://localhost:8080/api/vk/callback/interactionId2",
			},
			statusCode:   400,
			bodyContains: []string{"interaction does not exist"},
		},
		{
			name: "handleMessage[message_new]: ok",
			args: args{
				interactionId: "interactionId",
				reqBody:       []byte(`{"type":"message_new","event":"test_event_id","v":"1.0","group_id":12345,"object":{"message":{"from_id":1,"text":"test text"}}}`),
				url:           "http://localhost:8080/api/vk/callback/interactionId",
			},
			statusCode:   200,
			bodyContains: []string{"ok"},
			queueContents: []entities.Message{{
				InteractionId: "interactionId",
				Type:          entities.MessageTypeNew,
				Text:          "test text",
				VkSenderId:    1,
			}},
		},
		{
			name: "handleMessage[message_edit]: ok",
			args: args{
				interactionId: "interactionId",
				reqBody:       []byte(`{"type":"message_edit","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1,"text":"test text"}}`),
				url:           "http://localhost:8080/api/vk/callback/interactionId",
			},
			statusCode:   200,
			bodyContains: []string{"ok"},
			queueContents: []entities.Message{{
				InteractionId: "interactionId",
				Type:          entities.MessageTypeEdit,
				Text:          "test text",
				VkSenderId:    1,
			}},
		},
		{
			name: "handleMessage[message_reply]: ok",
			args: args{
				interactionId: "interactionId",
				reqBody:       []byte(`{"type":"message_reply","event":"test_event_id","v":"1.0","group_id":12345, "object":{"from_id":1,"text":"test text"}}`),
				url:           "http://localhost:8080/api/vk/callback/interactionId",
			},
			statusCode:   200,
			bodyContains: []string{"ok"},
			queueContents: []entities.Message{{
				InteractionId: "interactionId",
				Type:          entities.MessageTypeReply,
				Text:          "test text",
				VkSenderId:    1,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, _, client := setup(t, tt.args.interactionId, tt.args.confirmationString)
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

func setup(t *testing.T, interactionId, confirmationString string) (*queue.Queue[entities.Message], *bytes.Buffer, *http.Client) {
	t.Helper()
	q := queue.NewQueue[entities.Message]()
	buf := new(bytes.Buffer)
	log := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{}))
	repo := repository.NewStubRepo(interactionId, confirmationString, 321, 123)
	h := &handlers.Handlers{Metrics: &handlerStub{}, VkCallbackHandler: vk_callback_handler.New(q, repo, log)}

	listener := fasthttputil.NewInmemoryListener()
	p := gomonkey.ApplyFunc(net.Listen, func(network string, address string) (net.Listener, error) { return listener, nil })
	t.Cleanup(p.Reset)
	s := New(&Config{Host: "localhost", Port: 1337}, h, log)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = s.Run(ctx) }()

	client := &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return listener.Dial()
		},
	}}

	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		t.Helper()
		t.Log(buf.String())
	})
	return q, buf, client
}
