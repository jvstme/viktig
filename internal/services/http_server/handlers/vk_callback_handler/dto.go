package vk_callback_handler

type typeDto struct {
	Type       string `json:"type"`
	EventId    string `json:"event_id"`
	ApiVersion string `json:"v"`
	GroupId    int    `json:"group_id"`
}

type newMessageDto struct {
	Object struct {
		Message vkMessage `json:"message"`
	} `json:"object"`
}

type editOrReplyMessageDto struct {
	Object vkMessage `json:"object"`
}

type vkMessage struct {
	SenderId int    `json:"from_id"`
	Text     string `json:"text"`
}
