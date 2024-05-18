package repository

import "fmt"

type stubRepo struct {
	interaction Interaction
}

func NewStubRepo(hookId, confirmationString string, tgChatId int) Repository {
	return &stubRepo{interaction: Interaction{
		id:                 hookId,
		ConfirmationString: confirmationString,
		TgChatId:           tgChatId,
	}}
}

func (s *stubRepo) GetInteraction(id string) (*Interaction, error) {
	if id == "" {
		return nil, fmt.Errorf("empty interactionId")
	}
	if s.interaction.id == id {
		return &s.interaction, nil
	}
	return nil, nil
}

func (s *stubRepo) ExistsInteraction(id string) bool {
	return s.interaction.id == id
}
