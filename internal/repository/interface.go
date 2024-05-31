package repository

import (
	"viktig/internal/entities"

	"github.com/google/uuid"
)

type Repository interface {
	StoreInteraction(interaction *entities.Interaction) error
	ExistsInteraction(id uuid.UUID) bool
	GetInteraction(id uuid.UUID) (*entities.Interaction, error)
	DeleteInteraction(id uuid.UUID) error
	ListInteractions(userId int) ([]*entities.Interaction, error)

	StoreIncompleteInteraction(interaction *entities.IncompleteInteraction) error
	UpdateIncompleteInteraction(interaction *entities.IncompleteInteraction) error
	GetIncompleteInteraction(userId int) (*entities.IncompleteInteraction, error)
	DeleteIncompleteInteraction(userId int) error

	StoreUser(user *entities.User) error
	GetUser(id int) (*entities.User, error)
	DeleteUser(id int) error
}
