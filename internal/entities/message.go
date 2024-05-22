package entities

type Message struct {
	InteractionId string // same as hookId from vk
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
