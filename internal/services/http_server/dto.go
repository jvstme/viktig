package http_server

type typeDto struct {
	Type string `json:"type"`
}

type challengeDto struct {
	Type    string `json:"type"`
	GroupId int    `json:"group_id"`
}

type newMessageDto struct {
	Type    string `json:"type"`
	EventId int    `json:"event_id"`
	V       string `json:"v"`
	Object  struct {
		SenderId int    `json:"from_id"`
		Text     string `json:"text"`
	} `json:"object"`
	GroupId int `json:"group_id"`
}
