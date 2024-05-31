package in_memory_repo

import (
	"fmt"
	"viktig/internal/entities"
	"viktig/internal/repository"

	"github.com/google/uuid"
)

type inMemoryRepo struct {
	users                  map[int]*entities.User
	interactions           map[string]*entities.Interaction
	incompleteInteractions map[int]*entities.IncompleteInteraction
}

func New() repository.Repository {
	return &inMemoryRepo{
		users:                  make(map[int]*entities.User),
		interactions:           make(map[string]*entities.Interaction),
		incompleteInteractions: make(map[int]*entities.IncompleteInteraction),
	}
}

func (r *inMemoryRepo) GetInteraction(id uuid.UUID) (*entities.Interaction, error) {
	interaction, ok := r.interactions[id.String()]
	if !ok {
		return nil, fmt.Errorf("interaction not found")
	}
	return interaction, nil
}

func (r *inMemoryRepo) ExistsInteraction(id uuid.UUID) bool {
	return r.interactions[id.String()] != nil
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

func (r *inMemoryRepo) DeleteInteraction(id uuid.UUID) error {
	delete(r.interactions, id.String())
	return nil
}

func (r *inMemoryRepo) ListInteractions(userId int) ([]*entities.Interaction, error) {
	if _, ok := r.users[userId]; !ok {
		return nil, fmt.Errorf("user not found")
	}
	var interactions []*entities.Interaction
	for _, interaction := range r.interactions {
		if interaction.UserId == userId {
			interactions = append(interactions, interaction)
		}
	}
	return interactions, nil
}

func (r *inMemoryRepo) StoreIncompleteInteraction(interaction *entities.IncompleteInteraction) error {
	if interaction == nil {
		return nil
	}
	if _, ok := r.users[interaction.UserId]; !ok {
		return fmt.Errorf("user not found")
	}
	r.incompleteInteractions[interaction.UserId] = interaction
	return nil
}

func (r *inMemoryRepo) UpdateIncompleteInteraction(interaction *entities.IncompleteInteraction) error {
	return r.StoreIncompleteInteraction(interaction)
}

func (r *inMemoryRepo) GetIncompleteInteraction(userId int) (*entities.IncompleteInteraction, error) {
	interaction, ok := r.incompleteInteractions[userId]
	if !ok {
		return nil, fmt.Errorf("interaction not found")
	}
	return interaction, nil
}

func (r *inMemoryRepo) DeleteIncompleteInteraction(userId int) error {
	delete(r.incompleteInteractions, userId)
	return nil
}

func (r *inMemoryRepo) StoreUser(user *entities.User) error {
	if user == nil {
		return nil
	}
	r.users[user.Id] = user
	return nil
}

func (r *inMemoryRepo) GetUser(id int) (*entities.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (r *inMemoryRepo) DeleteUser(id int) error {
	delete(r.users, id)
	return nil
}
