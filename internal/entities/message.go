package entities

type Message struct {
	Type       MessageType
	Text       string
	VkSenderId int
	VkSender   *VkUser
}

type MessageType int

const (
	MessageTypeNew MessageType = iota
	MessageTypeEdit
	MessageTypeReply
)
