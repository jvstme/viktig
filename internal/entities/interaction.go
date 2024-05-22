package entities

import "github.com/google/uuid"

type Interaction struct { // todo: better naming for an item of forwarding
	Id                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	UserId             int       `json:"user_id"` // same as tg user id
	ConfirmationString string    `json:"confirmation_string"`
	TgChatId           int       `json:"tg_chat_id"`
}
