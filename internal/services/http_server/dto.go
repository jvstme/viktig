package http_server

type typeDto struct {
	Type string `json:"type"`
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
