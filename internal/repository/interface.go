package repository

import "viktig/internal/entities"

type Repository interface {
	NewInteraction(userId, tgChatId int, confirmationString string) (*entities.Interaction, error)
	ExistsInteraction(id string) bool
	GetInteraction(id string) (*entities.Interaction, error)

	NewUser(userId int) (*entities.User, error)
}
