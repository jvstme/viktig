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
	switch request.Action {
	case actionNewUser:
		if err = h.handleNewUser(ctx, request.Data.(*newUserRequestData)); err != nil {
			return
		}
	case actionGetUser:
		if err = h.handleGetUser(ctx, request.Data.(*getOrDeleteUserRequestData)); err != nil {
			return
		}
	case actionDeleteUser:
		if err = h.handleDeleteUser(ctx, request.Data.(*getOrDeleteUserRequestData)); err != nil {
			return
		}
	case actionNewInteraction:
		if err = h.handleNewInteraction(ctx, request.Data.(*newInteractionRequestData)); err != nil {
			return
		}
	case actionGetInteraction:
		if err = h.handleGetInteraction(ctx, request.Data.(*getOrDeleteInteractionRequestData)); err != nil {
			return
		}
	case actionDeleteInteraction:
		if err = h.handleDeleteInteraction(ctx, request.Data.(*getOrDeleteInteractionRequestData)); err != nil {
			return
		}
	}
}

func (h *debugHandler) handleNewInteraction(ctx *fasthttp.RequestCtx, data *newInteractionRequestData) error {
	if data == nil {
		return fmt.Errorf("request data is nil")
	}

	interaction := &entities.Interaction{
		Id:                 uuid.New(),
		Name:               data.InteractionName,
		UserId:             data.UserId,
		ConfirmationString: data.ConfirmationString,
		TgChatId:           data.TgChatId,
	}
	err := h.repo.StoreInteraction(interaction)
	if err != nil {
		return err
	}

	response := &newInteractionResponseDto{
		CallbackUrl: fmt.Sprintf("%s/api/vk/callback/%s", h.host, interaction.Id),
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

func (h *debugHandler) handleGetInteraction(ctx *fasthttp.RequestCtx, data *getOrDeleteInteractionRequestData) error {
	if data == nil {
		return fmt.Errorf("request data is nil")
	}

	interaction, err := h.repo.GetInteraction(data.Id)
	if err != nil {
		return err
	}

	bytes, err := jsoniter.Marshal(interaction)
	if err != nil {
		return err
	}
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("application/json")
	ctx.Response.SetBody(bytes)
	return nil
}

func (h *debugHandler) handleDeleteInteraction(ctx *fasthttp.RequestCtx, data *getOrDeleteInteractionRequestData) error {
	if data == nil {
		return fmt.Errorf("request data is nil")
	}

	interaction, err := h.repo.GetInteraction(data.Id)
	if err != nil {
		return err
	}
	if err = h.repo.DeleteInteraction(data.Id); err != nil {
		return err
	}

	bytes, err := jsoniter.Marshal(interaction)
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

func (h *debugHandler) handleGetUser(ctx *fasthttp.RequestCtx, data *getOrDeleteUserRequestData) error {
	if data == nil {
		return fmt.Errorf("request data is nil")
	}

	user, err := h.repo.GetUser(data.Id)
	if err != nil {
		return err
	}

	bytes, err := jsoniter.Marshal(user)
	if err != nil {
		return err
	}
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("application/json")
	ctx.Response.SetBody(bytes)
	return nil
}

func (h *debugHandler) handleDeleteUser(ctx *fasthttp.RequestCtx, data *getOrDeleteUserRequestData) error {
	if data == nil {
		return fmt.Errorf("request data is nil")
	}

	user, err := h.repo.GetUser(data.Id)
	if err != nil {
		return err
	}
	if err = h.repo.DeleteUser(data.Id); err != nil {
		return err
	}

	bytes, err := jsoniter.Marshal(user)
	if err != nil {
		return err
	}
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("application/json")
	ctx.Response.SetBody(bytes)
	return nil
}
