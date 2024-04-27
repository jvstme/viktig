package http_server

type typeDto struct {
	Type string `json:"type"`
}

type newMessageDto struct {
	Object struct {
		Message struct {
			SenderId int    `json:"from_id"`
			Text     string `json:"text"`
		} `json:"message"`
	} `json:"object"`
}
