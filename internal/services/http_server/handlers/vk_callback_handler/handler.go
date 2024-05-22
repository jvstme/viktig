package vk_callback_handler

import (
	"errors"
	"fmt"
	"log/slog"

	"viktig/internal/entities"
	"viktig/internal/metrics"
	"viktig/internal/queue"
	"viktig/internal/repository"
	"viktig/internal/services/http_server/handlers"

	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
)

const (
	responseBodyOk = "ok"

	InteractionIdKey = "interactionId"
)

type vkCallbackHandler struct {
	repo repository.Repository
	q    *queue.Queue[entities.Message]
	l    *slog.Logger
}

func New(q *queue.Queue[entities.Message], repo repository.Repository, l *slog.Logger) handlers.Handler {
	if q == nil || repo == nil {
		return nil
	}
	if l == nil {
		l = slog.Default()
	}

	return &vkCallbackHandler{q: q, repo: repo, l: l.With("name", "VkCallbackHandler")}
}

func (h *vkCallbackHandler) Handle(ctx *fasthttp.RequestCtx) {
	var err error
	defer func() {
		if err != nil {
			h.l.Error(fmt.Sprintf("error handling request: %+v", err))
			ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		}
	}()

	dto := &vkCallbackDto{}
	if err = jsoniter.Unmarshal(ctx.Request.Body(), dto); err != nil {
		err = fmt.Errorf("json unmarshal error: %w", err)
		return
	}

	h.l.Info("received vk event", "type", dto.Type, "id", dto.EventId, "groupId", dto.GroupId, "apiVersion", dto.ApiVersion)
	metrics.VKEventsReceived.With(prometheus.Labels{"type": string(dto.Type)}).Inc()

	switch dto.Type {
	case messageTypeChallenge:
		err = h.handleChallenge(ctx)
		return
	case messageTypeNew, messageTypeEdit, messageTypeReply:
		err = h.handleMessage(ctx, dto.Payload)
		return
	default:
		text := fmt.Sprintf("unsupported message type: %s", dto.Type)
		h.l.Warn(text, "messageType", dto.Type)
		ctx.Error(text, fasthttp.StatusBadRequest)
	}
}

func (h *vkCallbackHandler) handleChallenge(ctx *fasthttp.RequestCtx) error {
	interactionId, err := getInteractionId(ctx)
	if err != nil {
		return err
	}

	interaction, err := h.repo.GetInteraction(interactionId)
	if err != nil {
		return err
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("text/plain")
	ctx.Response.SetBody([]byte(interaction.ConfirmationString))
	return nil
}

func (h *vkCallbackHandler) handleMessage(ctx *fasthttp.RequestCtx, data *vkMessageData) error {
	interactionId, err := getInteractionId(ctx)
	if err != nil {
		return err
	}

	if !h.repo.ExistsInteraction(interactionId) {
		return fmt.Errorf("interaction does not exist")
	}

	//todo: put blocks. add timeout
	h.q.Put(entities.Message{
		InteractionId: interactionId,
		Type:          ForwardedMessageTypes[data.MessageType],
		Text:          data.Text,
		VkSenderId:    data.SenderId,
	})

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("text/plain")
	ctx.Response.SetBody([]byte(responseBodyOk))
	return nil
}

func getInteractionId(ctx *fasthttp.RequestCtx) (uuid.UUID, error) {
	strInteractionId, ok := ctx.UserValue(InteractionIdKey).(string)
	if !ok || strInteractionId == "" {
		// should be impossible but still check
		return uuid.Nil, errors.New("invalid interactionId")
	}
	interactionId, err := uuid.Parse(strInteractionId)
	if err != nil {
		return uuid.Nil, errors.New("invalid interactionId")
	}
	return interactionId, nil
}
