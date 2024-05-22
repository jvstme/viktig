package debug_handler

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
)

type Action string

const (
	actionNewInteraction Action = "new_interaction"
	actionNewUser        Action = "new_user"
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
	case actionNewInteraction:
		wrapper := &struct {
			Data *newInteractionRequestData `json:"data"`
		}{}
		if err := jsoniter.Unmarshal(bytes, wrapper); err != nil {
			return err
		}
		d.Data = wrapper.Data
	case actionNewUser:
		wrapper := &struct {
			Data *newUserRequestData `json:"data"`
		}{}
		if err := jsoniter.Unmarshal(bytes, wrapper); err != nil {
			return err
		}
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
	ConfirmationString string `json:"confirmation_string" validate:"required"`
	TgChatId           int    `json:"tg_chat_id" validate:"required"`
}

type newInteractionResponseDto struct {
	CallbackUrl string `json:"callback_url"`
}

type newUserRequestData struct {
	UserId int `json:"user_id" validate:"required"`
}
