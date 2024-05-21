package debug_handler

import (
	"fmt"

	"viktig/internal/repository"
	"viktig/internal/services/http_server/handlers"

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
		return
	}
	switch data := request.Data.(type) {
	case *registerInteractionRequestData:
		if err = h.handleRegisterInteraction(ctx, data); err != nil {
			return
		}
	case *newUserRequestData:
		if err = h.handleNewUser(ctx, data); err != nil {
			return
		}
	}
}

func (h *debugHandler) handleRegisterInteraction(ctx *fasthttp.RequestCtx, data *registerInteractionRequestData) error {
	interaction, err := h.repo.NewInteraction(data.UserId, data.TgChatId, data.ConfirmationString)
	if err != nil {
		return err
	}

	response := &registerInteractionResponseDto{
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
	_, err := h.repo.NewUser(data.UserId)
	if err != nil {
		return err
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("application/json")
	ctx.Response.SetBody([]byte("ok"))
	return nil
}
