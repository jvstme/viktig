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

func (m *Message) IsFromUser() bool {
	return m.VkSenderId > 0
}
