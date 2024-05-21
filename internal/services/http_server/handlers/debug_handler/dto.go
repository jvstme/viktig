package debug_handler

import jsoniter "github.com/json-iterator/go"

type Action string

const (
	actionRegister Action = "register"
	actionNewUser  Action = "new_user"
)

type debugRequestDto struct {
	Action Action      `json:"action" validate:"required"` //todo: check validation really executes
	Data   interface{} `json:"-"`                          // filled in custom json unmarshal
}

func (d *debugRequestDto) UnmarshalJSON(bytes []byte) error {
	// unmarshal all certain fields
	err := jsoniter.Unmarshal(bytes, d)
	if err != nil {
		return err
	}

	switch d.Action {
	case actionRegister:
		wrapper := &struct {
			Data *registerInteractionRequestData `json:"data"`
		}{}
		if err = jsoniter.Unmarshal(bytes, wrapper); err != nil {
			return err
		}
		d.Data = wrapper.Data
	case actionNewUser:
		wrapper := &struct {
			Data *newUserRequestData `json:"data"`
		}{}
		if err = jsoniter.Unmarshal(bytes, wrapper); err != nil {
			return err
		}
		d.Data = wrapper.Data
	default:
		return nil
	}

	return nil
}

type registerInteractionRequestData struct {
	UserId             int    `json:"user_id"`
	ConfirmationString string `json:"confirmation_string" validate:"required"`
	TgChatId           int    `json:"tg_chat_id" validate:"required"`
}

type registerInteractionResponseDto struct {
	CallbackUrl string `json:"callback_url"`
}

type newUserRequestData struct {
	UserId int `json:"user_id"`
}
