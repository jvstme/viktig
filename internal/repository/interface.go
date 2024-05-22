package repository

type Interaction struct { // todo: better naming for an item of forwarding
	id                 string
	ConfirmationString string
	TgChatId           int
}

type Repository interface {
	GetInteraction(id string) (*Interaction, error)
	ExistsInteraction(id string) bool
}
