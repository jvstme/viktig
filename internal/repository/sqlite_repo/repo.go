package sqlite_repo

import (
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"viktig/internal/entities"
	"viktig/internal/repository"
)

type sqliteRepo struct {
	db *gorm.DB
}

func New(cfg *repository.Config) repository.Repository {
	db, err := gorm.Open(sqlite.Open(cfg.Dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		panic(err)
	}
	if err = db.AutoMigrate(&entities.User{}, &entities.Interaction{}); err != nil {
		panic(err)
	}
	return &sqliteRepo{
		db: db,
	}
}

func (r *sqliteRepo) StoreInteraction(interaction *entities.Interaction) error {
	result := r.db.Create(interaction)
	return result.Error
}

func (r *sqliteRepo) ExistsInteraction(id uuid.UUID) bool {
	interaction := &entities.Interaction{}
	result := r.db.First(interaction, id)
	return result.Error == nil
}

func (r *sqliteRepo) GetInteraction(id uuid.UUID) (*entities.Interaction, error) {
	interaction := &entities.Interaction{}
	result := r.db.First(interaction, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return interaction, nil
}

func (r *sqliteRepo) DeleteInteraction(id uuid.UUID) error {
	interaction := &entities.Interaction{}
	result := r.db.Delete(interaction, id)
	return result.Error
}

func (r *sqliteRepo) StoreUser(user *entities.User) error {
	result := r.db.Create(user)
	return result.Error
}

func (r *sqliteRepo) GetUser(id int) (*entities.User, error) {
	user := &entities.User{}
	result := r.db.First(user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (r *sqliteRepo) DeleteUser(id int) error {
	user := &entities.User{}
	result := r.db.Delete(user, id)
	return result.Error
}
