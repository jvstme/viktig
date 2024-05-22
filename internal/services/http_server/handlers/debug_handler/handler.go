package debug_handler

import (
	"fmt"
	"viktig/internal/entities"
	"viktig/internal/repository"
	"viktig/internal/services/http_server/handlers"

	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

type debugHandler struct {
	repo repository.Repository
	host string
}

func New(host string, repo repository.Repository) handlers.Handler {
	return &debugHandler{host: host, repo: repo}
}

func (h *debugHandler) Handle(ctx *fasthttp.RequestCtx) {
	var err error
	defer func() {
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		}
	}()

	request := &debugRequestDto{}
	if err = jsoniter.Unmarshal(ctx.Request.Body(), request); err != nil {
		err = fmt.Errorf("json unmarshal error: %w", err)
		return
	}
	switch data := request.Data.(type) {
	case *newInteractionRequestData:
		if err = h.handleRegisterInteraction(ctx, data); err != nil {
			return
		}
	case *newUserRequestData:
		if err = h.handleNewUser(ctx, data); err != nil {
			return
		}
	}
}

func (h *debugHandler) handleRegisterInteraction(ctx *fasthttp.RequestCtx, data *newInteractionRequestData) error {
	if data == nil {
		return fmt.Errorf("request data is nil")
	}

	interaction := &entities.Interaction{
		Id:                 uuid.New(),
		UserId:             data.UserId,
		ConfirmationString: data.ConfirmationString,
		TgChatId:           data.TgChatId,
	}
	err := h.repo.StoreInteraction(interaction)
	if err != nil {
		return err
	}

	response := &newInteractionResponseDto{
		CallbackUrl: fmt.Sprintf("%s/callback/%s", h.host, interaction.Id),
	}
	bytes, err := jsoniter.Marshal(response)
	if err != nil {
		return err
	}
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("application/json")
	ctx.Response.SetBody(bytes)
	return nil
}

func (h *debugHandler) handleNewUser(ctx *fasthttp.RequestCtx, data *newUserRequestData) error {
	if data == nil {
		return fmt.Errorf("request data is nil")
	}

	err := h.repo.StoreUser(&entities.User{Id: data.UserId})
	if err != nil {
		return err
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("application/json")
	ctx.Response.SetBody([]byte("ok"))
	return nil
}
