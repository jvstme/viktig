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

type Forwarder struct {
	tgToken  string
	tgChatId int
	q        *queue.Queue[entities.Message]
}

func New(cfg *Config, queue *queue.Queue[entities.Message]) *Forwarder {
	return &Forwarder{
		tgToken:  cfg.TgConfig.Token,
		tgChatId: cfg.TgConfig.ChatId,
		q:        queue,
	}
}

func (f *Forwarder) Run(ctx context.Context) error {
	botSettings := tele.Settings{Token: f.tgToken}
	bot, err := tele.NewBot(botSettings)
	if err != nil {
		return err
	}

	slog.Info("forwarder is ready", "username", bot.Me.Username)

	for {
		select {
		case message := <-f.q.AsChan():
			sentMessage, err := bot.Send(
				tele.ChatID(f.tgChatId),
				render(message),
				tele.ModeHTML,
				tele.NoPreview,
			)
			if err != nil {
				slog.Error(err.Error())
			} else {
				slog.Info(
					"sent telegram message",
					"id", sentMessage.ID,
					"chatId", sentMessage.Chat.ID,
				)
				metrics.MessagesForwarded.Inc()
			}
		case <-ctx.Done():
			slog.Info("stopping forwarder service")
			return nil
		}
	}
}

func render(message entities.Message) string {
	entitySlug := "id"
	entityId := message.VkSenderId
	if message.VkSenderId < 0 {
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
