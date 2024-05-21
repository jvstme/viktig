package repository

import (
	"fmt"
	"github.com/hashicorp/go-uuid"
	"viktig/internal/entities"
)

type stubRepo struct {
	interaction entities.Interaction
	user        entities.User
}

func NewStubRepo(interactionId, confirmationString string, userId, tgChatId int) Repository {
	return &stubRepo{
		interaction: entities.Interaction{
			Id:                 interactionId,
			ConfirmationString: confirmationString,
			TgChatId:           tgChatId,
		},
		user: entities.User{Id: userId},
	}
}

func (s *stubRepo) GetInteraction(id string) (*entities.Interaction, error) {
	if id == "" {
		return nil, fmt.Errorf("empty interactionId")
	}
	if s.interaction.Id == id {
		return &s.interaction, nil
	}
	return nil, nil
}

func (s *stubRepo) ExistsInteraction(id string) bool {
	return s.interaction.Id == id
}

func (s *stubRepo) NewInteraction(userId, tgChatId int, confirmationString string) (*entities.Interaction, error) {
	id, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}
	s.interaction = entities.Interaction{
		Id:                 id,
		UserId:             userId,
		ConfirmationString: confirmationString,
		TgChatId:           tgChatId,
	}
	return &s.interaction, nil
}

func (s *stubRepo) NewUser(userId int) (*entities.User, error) {
	s.user = entities.User{Id: userId}
	return &s.user, nil
}
