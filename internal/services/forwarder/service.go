package forwarder

import (
	"context"
	"fmt"
	"html"
	"log/slog"

	"viktig/internal/entities"
	"viktig/internal/metrics"
	"viktig/internal/queue"
	"viktig/internal/repository"

	tele "gopkg.in/telebot.v3"
)

var messageTypeIcons = map[entities.MessageType]string{
	entities.MessageTypeNew:   "💬",
	entities.MessageTypeEdit:  "✏️",
	entities.MessageTypeReply: "↩️",
}

type Service struct {
	tgBotToken string
	q          *queue.Queue[entities.Message]
	repo       repository.Repository
	l          *slog.Logger
}

func New(cfg *Config, queue *queue.Queue[entities.Message], repo repository.Repository, l *slog.Logger) *Service {
	return &Service{
		tgBotToken: cfg.BotToken,
		q:          queue,
		repo:       repo,
		l:          l.With("name", "ForwarderService"),
	}
}

func (s *Service) Run(ctx context.Context) error {
	botSettings := tele.Settings{Token: s.tgBotToken}
	bot, err := tele.NewBot(botSettings)
	if err != nil {
		return fmt.Errorf("telebot error: %w", err)
	}

	s.l.Info("forwarder is ready", "username", bot.Me.Username)

	for {
		select {
		case <-ctx.Done():
			s.l.Info("stopping forwarder service")
			return nil
		case message := <-s.q.AsChan():
			l := s.l.With("interactionId", message.InteractionId, "fromVkSenderId", message.VkSenderId)
			interaction, err := s.repo.GetInteraction(message.InteractionId)
			if err != nil {
				l.Error(fmt.Errorf("get interaction failed: %w", err).Error())
				continue
			}
			if interaction == nil {
				l.Error("interaction not found", "interactionId", message.InteractionId)
				continue
			}

			l = l.With("toTgChatId", interaction.TgChatId)
			sentMessage, err := bot.Send(
				tele.ChatID(interaction.TgChatId),
				render(message),
				tele.ModeHTML,
				tele.NoPreview,
			)
			if err != nil {
				l.Error(fmt.Errorf("error sending tg message: %w", err).Error())
				continue
			}

			l.Info("sent telegram message", "sentMessageId", sentMessage.ID)
			metrics.MessagesForwarded.Inc()
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

	return fmt.Sprintf(
		"👤 <a href=\"https://vk.com/%s%d\">%d</a>\n%s %s",
		entitySlug,
		entityId,
		entityId,
		messageTypeIcons[message.Type],
		html.EscapeString(message.Text),
	)
}
