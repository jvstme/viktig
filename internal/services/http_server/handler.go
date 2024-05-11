package http_server

import (
	"errors"
	"fmt"
	"log/slog"
	"viktig/internal/entities"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

const (
	messageTypeChallenge = "confirmation"

	responseBodyOk = "ok"
)

var forwardedMessageTypes = map[string]entities.MessageType{
	"message_new":   entities.MessageTypeNew,
	"message_edit":  entities.MessageTypeEdit,
	"message_reply": entities.MessageTypeReply,
}

func (s *HttpServer) vkHandler(ctx *fasthttp.RequestCtx) {
	var err error
	defer func() {
		if err != nil {
			slog.Error(fmt.Sprintf("error handling request: %+v", err))
			ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		}
	}()

	dto := &typeDto{}
	if err = jsoniter.Unmarshal(ctx.Request.Body(), dto); err != nil {
		return
	}

	slog.Info(
		"received vk event",
		"type", dto.Type,
		"id", dto.EventId,
		"groupId", dto.GroupId,
		"apiVersion", dto.ApiVersion,
	)

	if dto.Type == messageTypeChallenge {
		err = s.handleChallenge(ctx)
	} else if messageType, ok := forwardedMessageTypes[dto.Type]; ok {
		err = s.handleMessage(ctx, messageType)
	} else {
		text := fmt.Sprintf("unsupported message type: %s", dto.Type)
		slog.Warn(text, "messageType", dto.Type)
		ctx.Error(text, fasthttp.StatusBadRequest)
	}
}

func (s *HttpServer) handleChallenge(ctx *fasthttp.RequestCtx) error {
	hookId, ok := ctx.UserValue(hookIdKey).(string)
	if !ok || hookId == "" {
		return errors.New("invalid hookId")
	}

	// todo: get confirmation string from hookId from DB

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("text/plain")
	ctx.Response.SetBody([]byte(s.ConfirmationString))

	return nil
}

func (s *HttpServer) handleMessage(ctx *fasthttp.RequestCtx, messageType entities.MessageType) error {

	hookId, ok := ctx.UserValue(hookIdKey).(string)
	if !ok || hookId == "" {
		return errors.New("invalid hookId")
	}

	// todo: enqueue message with hook

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

	s.q.Put(entities.Message{
		Type:       messageType,
		Text:       message.Text,
		VkSenderId: message.SenderId,
	})

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("text/plain")
	ctx.Response.SetBody([]byte(responseBodyOk))

	return nil
}
