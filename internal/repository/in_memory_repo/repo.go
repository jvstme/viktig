package in_memory_repo

import (
	"fmt"
	"github.com/google/uuid"
	"viktig/internal/entities"
	"viktig/internal/repository"
)

type inMemoryRepo struct {
	users        map[int]*entities.User
	interactions map[string]*entities.Interaction
}

func New() repository.Repository {
	return &inMemoryRepo{
		users:        make(map[int]*entities.User),
		interactions: make(map[string]*entities.Interaction),
	}
}

func (r *inMemoryRepo) GetInteraction(id uuid.UUID) (*entities.Interaction, error) {
	strId := id.String()
	interaction, ok := r.interactions[strId]
	if !ok {
		return nil, fmt.Errorf("interaction not found")
	}
	return interaction, nil
}

func (r *inMemoryRepo) ExistsInteraction(id uuid.UUID) bool {
	strId := id.String()
	return r.interactions[strId] != nil
}

func (r *inMemoryRepo) StoreInteraction(interaction *entities.Interaction) error {
	if interaction == nil {
		return nil
	}
	if _, ok := r.users[interaction.UserId]; !ok {
		return fmt.Errorf("user not found")
	}
	r.interactions[interaction.Id.String()] = interaction
	return nil
}

func (r *inMemoryRepo) StoreUser(user *entities.User) error {
	if user == nil {
		return nil
	}
	r.users[user.Id] = user
	return nil
}
