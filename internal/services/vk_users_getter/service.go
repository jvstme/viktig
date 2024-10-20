package vk_users_getter

import (
	"context"
	"log/slog"

	"viktig/internal/entities"
	"viktig/internal/queue"

	"github.com/go-vk-api/vk"
)

type VkUsersGetter struct {
	apiToken string
	qi       *queue.Queue[entities.Message]
	qo       *queue.Queue[entities.Message]
	l        *slog.Logger
}

func New(
	apiToken string,
	inQueue *queue.Queue[entities.Message],
	outQueue *queue.Queue[entities.Message],
	l *slog.Logger,
) *VkUsersGetter {
	return &VkUsersGetter{
		apiToken: apiToken,
		qi:       inQueue,
		qo:       outQueue,
		l:        l.With("service", "VkUsersGetter"),
	}
}

func (s *VkUsersGetter) Run(ctx context.Context) error {
	client, err := vk.NewClientWithOptions(
		vk.WithToken(s.apiToken),
		withLang("ru"), // disables names transliteration
	)
	if err != nil {
		return err
	}
	if err = checkVKClient(client); err != nil {
		return err
	}
	s.l.Info("vkUsersGetter is ready")

	for {
		select {
		case <-ctx.Done():
			s.l.Info("stopping vkUsersGetter service")
			return nil
		case message := <-s.qi.AsChan():
			// Retrieve VK users based on the sender ID of the incoming message
			if message.IsFromUser() {
				var users []*entities.VkUser
				err := client.CallMethod("users.get", vk.RequestParams{"user_id": message.VkSenderId}, &users)
				if err != nil || len(users) != 1 {
					s.l.Error("error getting user info", "entries", len(users), "err", err)
				} else {
					message.VkSender = users[0]
				}
			}
			s.qo.Put(message)
		}
	}
}

func checkVKClient(client *vk.Client) error {
	var users []entities.VkUser
	if err := client.CallMethod("users.get", vk.RequestParams{}, &users); err != nil {
		return err
	}
	return nil
}

func withLang(lang string) vk.Option {
	return func(client *vk.Client) error {
		client.Lang = lang
		return nil
	}
}
