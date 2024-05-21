package entities

type Interaction struct { // todo: better naming for an item of forwarding
	Id                 string
	UserId             int // same as tg user id
	ConfirmationString string
	TgChatId           int
}
