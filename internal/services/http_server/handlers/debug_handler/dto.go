package debug_handler

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

type Action string

const (
	actionNewInteraction    Action = "new_interaction"
	actionGetInteraction    Action = "get_interaction"
	actionDeleteInteraction Action = "delete_interaction"
	actionNewUser           Action = "new_user"
	actionGetUser           Action = "get_user"
	actionDeleteUser        Action = "delete_user"
)

type actionDto struct {
	Action Action `json:"action" validate:"required"`
}

type debugRequestDto struct {
	actionDto
	Data interface{} `json:"-"`
}

func (d *debugRequestDto) UnmarshalJSON(bytes []byte) error {
	if err := jsoniter.Unmarshal(bytes, &d.actionDto); err != nil {
		return err
	}

	switch d.Action {
	case actionNewUser:
		wrapper := &struct {
			Data *newUserRequestData `json:"data"`
		}{}
		_ = jsoniter.Unmarshal(bytes, wrapper) // not the first unmarshal
		d.Data = wrapper.Data
	case actionDeleteUser, actionGetUser:
		wrapper := &struct {
			Data *getOrDeleteUserRequestData `json:"data"`
		}{}
		_ = jsoniter.Unmarshal(bytes, wrapper) // not the first unmarshal
		d.Data = wrapper.Data
	case actionNewInteraction:
		wrapper := &struct {
			Data *newInteractionRequestData `json:"data"`
		}{}
		_ = jsoniter.Unmarshal(bytes, wrapper) // not the first unmarshal
		d.Data = wrapper.Data
	case actionGetInteraction, actionDeleteInteraction:
		wrapper := &struct {
			Data *getOrDeleteInteractionRequestData `json:"data"`
		}{}
		_ = jsoniter.Unmarshal(bytes, wrapper) // not the first unmarshal
		d.Data = wrapper.Data
	}

	err := validator.New().Struct(d)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	return nil
}

type newInteractionRequestData struct {
	UserId             int    `json:"user_id" validate:"required"`
	InteractionName    string `json:"name" validate:"required"`
	ConfirmationString string `json:"confirmation_string" validate:"required"`
	TgChatId           int    `json:"tg_chat_id" validate:"required"`
}

type newInteractionResponseDto struct {
	CallbackUrl string `json:"callback_url"`
}

type getOrDeleteInteractionRequestData struct {
	Id uuid.UUID `json:"id" validate:"required"`
}

type newUserRequestData struct {
	UserId int `json:"id" validate:"required"`
}

type getOrDeleteUserRequestData struct {
	Id int `json:"id" validate:"required"`
}
