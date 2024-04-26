package forwarder

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"viktig/internal/core"

	tele "gopkg.in/telebot.v3"
)

type Forwarder struct {
	tgToken  string
	tgChatId int
}

func New(cfg *Config) *Forwarder {
	return &Forwarder{
		tgToken:  cfg.TgConfig.Token,
		tgChatId: cfg.TgConfig.ChatId,
	}
}

func (f *Forwarder) Run(ctx context.Context, messages chan core.Message) error {
	botSettings := tele.Settings{Token: f.tgToken}
	bot, err := tele.NewBot(botSettings)
	if err != nil {
		return err
	}

	for {
		select {
		case message := <-messages:
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

func render(message core.Message) string {
	return fmt.Sprintf(
		"👤 <a href=\"https://vk.com/id%d\">%d</a>\n💬 %s",
		message.VkSenderId,
		message.VkSenderId,
		html.EscapeString(message.Text),
	)
}
