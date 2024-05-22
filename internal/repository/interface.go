package repository

import (
	"github.com/google/uuid"
	"viktig/internal/entities"
)

type Repository interface {
	StoreInteraction(interaction *entities.Interaction) error
	ExistsInteraction(id uuid.UUID) bool
	GetInteraction(id uuid.UUID) (*entities.Interaction, error)
	DeleteInteraction(id uuid.UUID) error

	StoreUser(user *entities.User) error
	GetUser(id int) (*entities.User, error)
	DeleteUser(id int) error
}
