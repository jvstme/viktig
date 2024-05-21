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

	jsoniter "github.com/json-iterator/go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
)

const (
	messageTypeChallenge = "confirmation"

	responseBodyOk = "ok"

	HookIdKey = "community_hook_id"
)

var ForwardedMessageTypes = map[string]entities.MessageType{
	"message_new":   entities.MessageTypeNew,
	"message_edit":  entities.MessageTypeEdit,
	"message_reply": entities.MessageTypeReply,
}

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

	dto := &typeDto{}
	if err = jsoniter.Unmarshal(ctx.Request.Body(), dto); err != nil {
		err = fmt.Errorf("json unmarshal error: %w", err)
		return
	}

	h.l.Info("received vk event", "type", dto.Type, "id", dto.EventId, "groupId", dto.GroupId, "apiVersion", dto.ApiVersion)
	metrics.VKEventsReceived.With(prometheus.Labels{"type": dto.Type}).Inc()

	if dto.Type == messageTypeChallenge {
		err = h.handleChallenge(ctx)
		return
	}
	if messageType, ok := ForwardedMessageTypes[dto.Type]; ok {
		err = h.handleMessage(ctx, messageType)
		return
	}

	text := fmt.Sprintf("unsupported message type: %s", dto.Type)
	h.l.Warn(text, "messageType", dto.Type)
	ctx.Error(text, fasthttp.StatusBadRequest)
}

func (h *vkCallbackHandler) handleChallenge(ctx *fasthttp.RequestCtx) error {
	hookId, ok := ctx.UserValue(HookIdKey).(string)
	if !ok || hookId == "" {
		// should be impossible but still check
		return errors.New("invalid hookId")
	}

	interaction, err := h.repo.GetInteraction(hookId)
	if err != nil {
		return err
	}
	if interaction == nil {
		return errors.New("interaction not found")
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("text/plain")
	ctx.Response.SetBody([]byte(interaction.ConfirmationString))
	return nil
}

func (h *vkCallbackHandler) handleMessage(ctx *fasthttp.RequestCtx, messageType entities.MessageType) error {
	hookId, ok := ctx.UserValue(HookIdKey).(string)
	if !ok || hookId == "" {
		// should be impossible but still check
		return errors.New("invalid hookId")
	}

	var message *vkMessage
	if messageType == entities.MessageTypeNew {
		dto := &newMessageDto{}
		if err := jsoniter.Unmarshal(ctx.Request.Body(), dto); err != nil {
			return err
		}
		message = &dto.Object.Message
	} else {
		dto := &editOrReplyMessageDto{}
		if err := jsoniter.Unmarshal(ctx.Request.Body(), dto); err != nil {
			return err
		}
		message = &dto.Object
	}

	if !h.repo.ExistsInteraction(hookId) {
		return fmt.Errorf("interaction does not exist")
	}

	//todo: put blocks. add timeout
	h.q.Put(entities.Message{
		InteractionId: hookId,
		Type:          messageType,
		Text:          message.Text,
		VkSenderId:    message.SenderId,
	})

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("text/plain")
	ctx.Response.SetBody([]byte(responseBodyOk))
	return nil
}
