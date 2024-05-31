package forwarder

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"

	"viktig/internal/entities"
	"viktig/internal/queue"
	"viktig/internal/repository/in_memory_repo"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	tele "gopkg.in/telebot.v3"
)

func TestRender(t *testing.T) {
	t.Run("new message", func(t *testing.T) {
		message := entities.Message{
			Type:       entities.MessageTypeNew,
			Text:       "Hello",
			VkSenderId: 1234,
		}
		actual := render(message)
		expected := "👤 <a href=\"https://vk.com/id1234\">1234</a>\n💬 Hello"
		assert.Equal(t, expected, actual)
	})
	t.Run("edited message", func(t *testing.T) {
		message := entities.Message{
			Type:       entities.MessageTypeEdit,
			Text:       "Edit",
			VkSenderId: 1234,
		}
		actual := render(message)
		expected := "👤 <a href=\"https://vk.com/id1234\">1234</a>\n✏️ Edit"
		assert.Equal(t, expected, actual)
	})
	t.Run("edited by community message", func(t *testing.T) {
		message := entities.Message{
			Type:       entities.MessageTypeEdit,
			Text:       "Edit",
			VkSenderId: -123,
		}
		actual := render(message)
		expected := "👤 <a href=\"https://vk.com/club123\">123</a>\n✏️ Edit"
		assert.Equal(t, expected, actual)
	})
	t.Run("replied message", func(t *testing.T) {
		message := entities.Message{
			Type:       entities.MessageTypeReply,
			Text:       "Reply",
			VkSenderId: 4321,
		}
		actual := render(message)
		expected := "👤 <a href=\"https://vk.com/id4321\">4321</a>\n↩️ Reply"
		assert.Equal(t, expected, actual)
	})
	t.Run("escape HTML", func(t *testing.T) {
		message := entities.Message{
			Type:       entities.MessageTypeNew,
			Text:       "<a href=\"https://x.com\">&</a>",
			VkSenderId: 1,
		}
		actual := render(message)
		expected := "👤 <a href=\"https://vk.com/id1\">1</a>\n💬 &lt;a href=&#34;https://x.com&#34;&gt;&amp;&lt;/a&gt;"
		assert.Equal(t, expected, actual)
	})
}

func TestService(t *testing.T) {
	interactionId := uuid.New()
	interactionId2 := uuid.New()
	t.Run("stop", func(t *testing.T) {
		p := gomonkey.ApplyFunc(tele.NewBot, func(_ tele.Settings) (*tele.Bot, error) {
			return &tele.Bot{Me: &tele.User{Username: "mock"}}, nil
		})
		defer p.Reset()
		_, _, s := setup(t, interactionId, "confirmationString", "token", 123)
		errCh := make(chan error)
		ctx, cancel := context.WithCancel(context.Background())
		go func() { errCh <- s.Run(ctx) }()
		cancel()

		assert.NoError(t, <-errCh)
	})
	t.Run("error get interaction", func(t *testing.T) {
		p := gomonkey.ApplyFunc(tele.NewBot, func(_ tele.Settings) (*tele.Bot, error) {
			return &tele.Bot{Me: &tele.User{Username: "mock"}}, nil
		})
		defer p.Reset()
		q, buf, s := setup(t, interactionId, "confirmationString", "token", 123)
		ctx, cancel := context.WithCancel(context.Background())
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() { defer wg.Done(); _ = s.Run(ctx) }()

		q.Put(entities.Message{
			InteractionId: interactionId2,
			Type:          entities.MessageTypeNew,
			Text:          "Hello",
			VkSenderId:    1234,
		})

		cancel()
		wg.Wait()

		logOutput := buf.String()
		assert.Contains(t, logOutput, `interaction not found`)
		assert.Contains(t, logOutput, fmt.Sprintf(`interactionId=%s`, interactionId2))
	})
	t.Run("send error", func(t *testing.T) {
		fakeBot := &tele.Bot{Me: &tele.User{Username: "mock"}}
		p := gomonkey.
			ApplyFunc(tele.NewBot, func(_ tele.Settings) (*tele.Bot, error) {
				return fakeBot, nil
			}).
			ApplyMethodFunc(fakeBot, "Send", func(to tele.Recipient, what interface{}, opts ...interface{}) (*tele.Message, error) {
				return nil, fmt.Errorf("error")
			})
		defer p.Reset()
		q, buf, s := setup(t, interactionId, "confirmationString", "token", 123)
		ctx, cancel := context.WithCancel(context.Background())
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() { defer wg.Done(); _ = s.Run(ctx) }()

		q.Put(entities.Message{
			InteractionId: interactionId,
			Type:          entities.MessageTypeNew,
			Text:          "Hello",
			VkSenderId:    1234,
		})

		cancel()
		wg.Wait()

		logOutput := buf.String()
		assert.Contains(t, logOutput, `error sending tg message: error`)
		assert.Contains(t, logOutput, fmt.Sprintf(`interactionId=%s`, interactionId))
		assert.Contains(t, logOutput, `fromVkSenderId=1234`)
		assert.Contains(t, logOutput, `toTgChatId=123`)
	})
	t.Run("ok", func(t *testing.T) {
		fakeBot := &tele.Bot{Me: &tele.User{Username: "mock"}}
		p := gomonkey.
			ApplyFunc(tele.NewBot, func(_ tele.Settings) (*tele.Bot, error) {
				return fakeBot, nil
			}).
			ApplyMethodFunc(fakeBot, "Send", func(_ tele.Recipient, _ interface{}, _ ...interface{}) (*tele.Message, error) {
				return &tele.Message{ID: 321}, nil
			})
		defer p.Reset()
		q, buf, s := setup(t, interactionId, "confirmationString", "token", 123)
		ctx, cancel := context.WithCancel(context.Background())
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() { defer wg.Done(); _ = s.Run(ctx) }()

		q.Put(entities.Message{
			InteractionId: interactionId,
			Type:          entities.MessageTypeNew,
			Text:          "Hello",
			VkSenderId:    1234,
		})

		cancel()
		wg.Wait()

		logOutput := buf.String()
		assert.Contains(t, logOutput, `sent telegram message`)
		assert.Contains(t, logOutput, `sentMessageId=321`)
		assert.Contains(t, logOutput, fmt.Sprintf(`interactionId=%s`, interactionId))
		assert.Contains(t, logOutput, `fromVkSenderId=1234`)
		assert.Contains(t, logOutput, `toTgChatId=123`)
	})
}

func setup(t *testing.T, interactionId uuid.UUID, confirmationString, botToken string, tgChatId int) (*queue.Queue[entities.Message], *bytes.Buffer, *Service) {
	t.Helper()
	q := queue.NewQueue[entities.Message]()

	buf := new(bytes.Buffer)
	log := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{}))

	repo := in_memory_repo.New()
	_ = repo.StoreUser(&entities.User{Id: 123})
	_ = repo.StoreInteraction(&entities.Interaction{
		Id:                 interactionId,
		UserId:             123,
		ConfirmationString: confirmationString,
		TgChatId:           tgChatId,
	})

	s := New(nil, q, repo, log)

	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		t.Helper()
		buf.Reset()
	})
	return q, buf, s
}
