package repository

import (
	"github.com/google/uuid"
	"viktig/internal/entities"
)

type Repository interface {
	StoreInteraction(interaction *entities.Interaction) error
	ExistsInteraction(id uuid.UUID) bool
	GetInteraction(id uuid.UUID) (*entities.Interaction, error)

	StoreUser(user *entities.User) error
}
