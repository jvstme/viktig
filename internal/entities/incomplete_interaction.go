package entities

type IncompleteInteraction struct {
	UserId             int     `gorm:"primaryKey;foreignKey:UserId"` // todo: no autoincrement, also in users table
	Name               *string `gorm:"type:varchar(255)"`
	TgChatId           *int
	ConfirmationString *string `gorm:"type:varchar(255)"`
}
