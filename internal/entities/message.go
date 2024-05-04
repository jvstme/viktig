package entities

type Message struct {
	Type       MessageType
	Text       string
	VkSenderId int
}

type MessageType int

const (
	MessageTypeNew MessageType = iota
	MessageTypeEdit
	MessageTypeReply
)
