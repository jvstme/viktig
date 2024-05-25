package entities

import (
	"github.com/google/uuid"
)

type Interaction struct { // todo: better naming for an item of forwarding
	Id                 uuid.UUID `json:"id" gorm:"primaryKey;type:uuid"`
	Name               string    `json:"name" gorm:"type:varchar(255);not null"`
	UserId             int       `json:"user_id" gorm:"foreignKey:UserId;not null"`
	ConfirmationString string    `json:"confirmation_string" gorm:"type:varchar(255);not null"`
	TgChatId           int       `json:"tg_chat_id" gorm:"not null"`
}
