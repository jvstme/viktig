package forwarder

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strconv"

	"viktig/internal/entities"
	"viktig/internal/metrics"
	"viktig/internal/queue"

	tele "gopkg.in/telebot.v3"
)

var messageTypeIcons = map[entities.MessageType]string{
	entities.MessageTypeNew:   "💬",
	entities.MessageTypeEdit:  "✏️",
	entities.MessageTypeReply: "↩️",
}

type Community struct {
	TgChatId int
}

type Forwarder struct {
	tgToken     string
	communities map[string]*Community
	q           *queue.Queue[entities.Message]
	l           *slog.Logger
}

func New(
	tgToken string,
	communities map[string]*Community,
	q *queue.Queue[entities.Message],
	l *slog.Logger,
) *Forwarder {
	return &Forwarder{
		tgToken:     tgToken,
		communities: communities,
		q:           q,
		l:           l.With("service", "Forwarder"),
	}
}

func (f *Forwarder) Run(ctx context.Context) error {
	botSettings := tele.Settings{Token: f.tgToken}
	bot, err := tele.NewBot(botSettings)
	if err != nil {
		return fmt.Errorf("telebot error: %w", err)
	}

	f.l.Info("forwarder is ready", "username", bot.Me.Username)

	for {
		select {
		case message := <-f.q.AsChan():
			community, ok := f.communities[message.HookId]
			if !ok {
				f.l.Error("hookId not found", "hookId", message.HookId)
				continue
			}
			sentMessage, err := bot.Send(
				tele.ChatID(community.TgChatId),
				render(message),
				tele.ModeHTML,
				tele.NoPreview,
			)
			if err != nil {
				f.l.Error("error sending telegram message", "err", err.Error())
			} else {
				f.l.Info(
					"sent telegram message",
					"id", sentMessage.ID,
					"chatId", sentMessage.Chat.ID,
				)
				metrics.MessagesForwarded.Inc()
			}
		case <-ctx.Done():
			f.l.Info("stopping forwarder service")
			return nil
		}
	}
}

func render(message entities.Message) string {
	entitySlug := "id"
	entityId := message.VkSenderId
	if !message.IsFromUser() {
		entitySlug = "club"
		entityId = -message.VkSenderId
	}
	userName := strconv.Itoa(entityId)
	if message.VkSender != nil {
		userName = message.VkSender.FirstName + " " + message.VkSender.LastName
	}

	return fmt.Sprintf(
		"👤 <a href=\"https://vk.com/%s%d\">%s</a>\n%s %s",
		entitySlug,
		entityId,
		userName,
		messageTypeIcons[message.Type],
		html.EscapeString(message.Text),
	)
}
