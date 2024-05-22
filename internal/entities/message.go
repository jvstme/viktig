package entities

import "github.com/google/uuid"

type Message struct {
	InteractionId uuid.UUID
	Type          MessageType
	Text          string
	VkSenderId    int
}

type MessageType int

const (
	MessageTypeNew MessageType = iota
	MessageTypeEdit
	MessageTypeReply
)
