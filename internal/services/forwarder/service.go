package forwarder

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"viktig/internal/entities"
	"viktig/internal/queue"

	tele "gopkg.in/telebot.v3"
)

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
				slog.Info("sent message", "id", sentMessage.ID)
			}
		case <-ctx.Done():
			slog.Info("stopping forwarder service...")
			return nil
		}
	}
}

func render(message entities.Message) string {
	return fmt.Sprintf(
		"👤 <a href=\"https://vk.com/id%d\">%d</a>\n💬 %s",
		message.VkSenderId,
		message.VkSenderId,
		html.EscapeString(message.Text),
	)
}
