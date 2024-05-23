package vk_callback_handler

import (
	"fmt"

	"viktig/internal/entities"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
)

type MessageType string

const (
	messageTypeChallenge MessageType = "confirmation"
	messageTypeNew                   = "message_new"
	messageTypeEdit                  = "message_edit"
	messageTypeReply                 = "message_reply"
)

var ForwardedMessageTypes = map[MessageType]entities.MessageType{
	messageTypeNew:   entities.MessageTypeNew,
	messageTypeEdit:  entities.MessageTypeEdit,
	messageTypeReply: entities.MessageTypeReply,
}

type typeDto struct {
	Type       MessageType `json:"type"`
	EventId    string      `json:"event_id"`
	ApiVersion string      `json:"v"`
	GroupId    int         `json:"group_id"`
}

type vkCallbackDto struct {
	typeDto
	Payload *vkMessageData `json:"-"`
}

func (d *vkCallbackDto) UnmarshalJSON(bytes []byte) error {
	if err := jsoniter.Unmarshal(bytes, &d.typeDto); err != nil {
		return err
	}

	switch d.Type {
	case messageTypeChallenge:
		// no payload
		return nil
	case messageTypeNew:
		wrapper := struct {
			Object struct {
				Message *vkMessageData `json:"message"`
			} `json:"object"`
		}{}
		_ = jsoniter.Unmarshal(bytes, &wrapper) // not the first unmarshal
		d.Payload = wrapper.Object.Message
	case messageTypeEdit, messageTypeReply:
		wrapper := struct {
			Message *vkMessageData `json:"object"`
		}{}
		_ = jsoniter.Unmarshal(bytes, &wrapper) // not the first unmarshal
		d.Payload = wrapper.Message
	default:
		// not supported type
		return nil
	}
	d.Payload.MessageType = d.Type

	if err := validator.New().Struct(d); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	return nil
}

type vkMessageData struct {
	MessageType MessageType `json:"-" validate:"required"`
	SenderId    int         `json:"from_id" validate:"required"`
	Text        string      `json:"text" validate:"required"`
}
