package sqlite_repo

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log/slog"
	"viktig/internal/entities"
	"viktig/internal/repository"
	"viktig/internal/repository/in_memory_repo"
)

type postgresRepo struct {
	db *gorm.DB
}

func New() repository.Repository {
	db, err := gorm.Open(sqlite.Open("./dev/dev.db"), &gorm.Config{})
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to connect to database use in-memory repo: %v", err))
		return in_memory_repo.New()
	}
	_ = db.AutoMigrate(&entities.User{}, &entities.Interaction{})
	_ = db.AutoMigrate(&entities.Interaction{})
	return &postgresRepo{
		db: db,
	}
}

func (r *postgresRepo) StoreInteraction(interaction *entities.Interaction) error {
	result := r.db.Create(interaction)
	return result.Error
}

func (r *postgresRepo) ExistsInteraction(id uuid.UUID) bool {
	interaction := &entities.Interaction{}
	result := r.db.First(interaction, id)
	return result.Error == nil
}

func (r *postgresRepo) GetInteraction(id uuid.UUID) (*entities.Interaction, error) {
	interaction := &entities.Interaction{}
	result := r.db.First(interaction, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return interaction, nil
}

func (r *postgresRepo) DeleteInteraction(id uuid.UUID) error {
	interaction := &entities.Interaction{}
	result := r.db.Delete(interaction, id)
	return result.Error
}

func (r *postgresRepo) StoreUser(user *entities.User) error {
	result := r.db.Create(user)
	return result.Error
}

func (r *postgresRepo) GetUser(id int) (*entities.User, error) {
	user := &entities.User{}
	result := r.db.First(user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (r *postgresRepo) DeleteUser(id int) error {
	user := &entities.User{}
	result := r.db.Delete(user, id)
	return result.Error
}
