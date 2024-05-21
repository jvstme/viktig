package in_memory_repo

import (
	"fmt"
	"github.com/hashicorp/go-uuid"
	"viktig/internal/entities"
	"viktig/internal/repository"
)

type inMemoryRepo struct {
	users        map[int]*entities.User
	interactions map[string]*entities.Interaction
}

func New() repository.Repository {
	return &inMemoryRepo{}
}

func (r *inMemoryRepo) GetInteraction(id string) (*entities.Interaction, error) {
	interaction, ok := r.interactions[id]
	if !ok {
		return nil, fmt.Errorf("interaction not found")
	}
	return interaction, nil
}

func (r *inMemoryRepo) ExistsInteraction(id string) bool {
	return r.interactions[id] != nil
}

func (r *inMemoryRepo) NewInteraction(userId, tgChatId int, confirmationString string) (*entities.Interaction, error) {
	if _, ok := r.users[userId]; !ok {
		return nil, fmt.Errorf("user not found")
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}
	interaction := &entities.Interaction{
		Id:                 id,
		UserId:             userId,
		ConfirmationString: confirmationString,
		TgChatId:           tgChatId,
	}
	r.interactions[id] = interaction
	return interaction, nil
}

func (r *inMemoryRepo) NewUser(userId int) (*entities.User, error) {
	if _, ok := r.users[userId]; ok {
		return nil, fmt.Errorf("user %d already exists", userId)
	}
	user := &entities.User{
		Id: userId,
	}
	r.users[userId] = user
	return user, nil
}
