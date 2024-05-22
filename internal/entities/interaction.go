package entities

import "github.com/google/uuid"

type Interaction struct { // todo: better naming for an item of forwarding
	Id                 uuid.UUID
	UserId             int // same as tg user id
	ConfirmationString string
	TgChatId           int
}
