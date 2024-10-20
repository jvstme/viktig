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

	"github.com/agiledragon/gomonkey/v2"
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
	t.Run("with sender name", func(t *testing.T) {
		message := entities.Message{
			Type:       entities.MessageTypeNew,
			Text:       "Hello",
			VkSenderId: 1234,
			VkSender:   &entities.VkUser{FirstName: "John", LastName: "Doe"},
		}
		actual := render(message)
		expected := "👤 <a href=\"https://vk.com/id1234\">John Doe</a>\n💬 Hello"
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
	t.Run("stop", func(t *testing.T) {
		p := gomonkey.ApplyFunc(tele.NewBot, func(_ tele.Settings) (*tele.Bot, error) {
			return &tele.Bot{Me: &tele.User{Username: "mock"}}, nil
		})
		defer p.Reset()
		_, _, s := setup(t, nil)
		errCh := make(chan error)
		ctx, cancel := context.WithCancel(context.Background())
		go func() { errCh <- s.Run(ctx) }()
		cancel()

		assert.NoError(t, <-errCh)
	})
	t.Run("stop on bot error", func(t *testing.T) {
		p := gomonkey.ApplyFunc(tele.NewBot, func(_ tele.Settings) (*tele.Bot, error) {
			return nil, fmt.Errorf("error")
		})
		defer p.Reset()
		_, _, s := setup(t, nil)
		errCh := make(chan error)
		ctx, cancel := context.WithCancel(context.Background())
		go func() { errCh <- s.Run(ctx) }()
		cancel()

		assert.EqualError(t, <-errCh, "telebot error: error")
	})
	t.Run("hookId not found", func(t *testing.T) {
		fakeBot := &tele.Bot{Me: &tele.User{Username: "mock"}}
		p := gomonkey.ApplyFunc(tele.NewBot, func(_ tele.Settings) (*tele.Bot, error) {
			return fakeBot, nil
		})
		defer p.Reset()
		q, buf, s := setup(t, map[string]*Community{"test-hook": {TgChatId: 4321}})
		ctx, cancel := context.WithCancel(context.Background())
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() { defer wg.Done(); _ = s.Run(ctx) }()

		q.Put(entities.Message{
			HookId:     "unknown-hook",
			Type:       entities.MessageTypeNew,
			Text:       "Hello",
			VkSenderId: 1234,
		})

		cancel()
		wg.Wait()

		logOutput := buf.String()
		assert.Contains(t, logOutput, "hookId not found")
		assert.Contains(t, logOutput, "hookId=unknown-hook")
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
		q, buf, s := setup(t, map[string]*Community{"test-hook": {TgChatId: 4321}})
		ctx, cancel := context.WithCancel(context.Background())
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() { defer wg.Done(); _ = s.Run(ctx) }()

		q.Put(entities.Message{
			HookId:     "test-hook",
			Type:       entities.MessageTypeNew,
			Text:       "Hello",
			VkSenderId: 1234,
		})

		cancel()
		wg.Wait()

		logOutput := buf.String()
		assert.Contains(t, logOutput, "error sending telegram message")
		assert.Contains(t, logOutput, "err=error")
	})
	t.Run("ok", func(t *testing.T) {
		fakeBot := &tele.Bot{Me: &tele.User{Username: "mock"}}
		p := gomonkey.
			ApplyFunc(tele.NewBot, func(_ tele.Settings) (*tele.Bot, error) {
				return fakeBot, nil
			}).
			ApplyMethodFunc(fakeBot, "Send", func(_ tele.Recipient, _ interface{}, _ ...interface{}) (*tele.Message, error) {
				return &tele.Message{ID: 321, Chat: &tele.Chat{ID: 4321}}, nil
			})
		defer p.Reset()
		q, buf, s := setup(t, map[string]*Community{"test-hook": {TgChatId: 4321}})
		ctx, cancel := context.WithCancel(context.Background())
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() { defer wg.Done(); _ = s.Run(ctx) }()

		q.Put(entities.Message{
			HookId:     "test-hook",
			Type:       entities.MessageTypeNew,
			Text:       "Hello",
			VkSenderId: 1234,
		})

		cancel()
		wg.Wait()

		logOutput := buf.String()
		assert.Contains(t, logOutput, "sent telegram message")
		assert.Contains(t, logOutput, "id=321")
		assert.Contains(t, logOutput, "chatId=4321")
	})
}

func setup(
	t *testing.T,
	communities map[string]*Community,
) (*queue.Queue[entities.Message], *bytes.Buffer, *Forwarder) {
	t.Helper()
	q := queue.NewQueue[entities.Message]()

	buf := new(bytes.Buffer)
	log := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{}))

	s := New("token", communities, q, log)

	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		t.Helper()
		buf.Reset()
	})
	return q, buf, s
}
