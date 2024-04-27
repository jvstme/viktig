package http_server

import (
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"log/slog"
	"viktig/internal/entities"
)

const (
	messageTypeChallenge  = "confirmation"
	messageTypeNewMessage = "message_new"

	responseBodyOk = "ok"
)

func (s *HttpServer) vkHandler(ctx *fasthttp.RequestCtx) {
	var err error
	defer func() {
		if err != nil {
			slog.Error(fmt.Sprintf("error handling request: %+v", err))
			ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		}
	}()

	var dto *typeDto
	if err = jsoniter.Unmarshal(ctx.Request.Body(), dto); err != nil {
		return
	}

	switch dto.Type {
	case messageTypeChallenge:
		err = s.handleChallenge(ctx)
	case messageTypeNewMessage:
		err = s.handleNewMessage(ctx)
	default:
		text := fmt.Sprintf("unsupported message type: %s", dto.Type)
		slog.Warn(text, "messageType", dto.Type)
		ctx.Error(text, fasthttp.StatusBadRequest)
	}
}

func (s *HttpServer) handleChallenge(ctx *fasthttp.RequestCtx) error {
	var dto *challengeDto
	if err := jsoniter.Unmarshal(ctx.Request.Body(), dto); err != nil {
		return err
	}

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

func (s *HttpServer) handleNewMessage(ctx *fasthttp.RequestCtx) error {
	var dto *newMessageDto
	if err := jsoniter.Unmarshal(ctx.Request.Body(), dto); err != nil {
		return err
	}

	hookId, ok := ctx.UserValue(hookIdKey).(string)
	if !ok || hookId == "" {
		return errors.New("invalid hookId")
	}

	// todo: enqueue message with hook

	s.q.Put(entities.Message{
		Text:       dto.Object.Text,
		VkSenderId: dto.Object.SenderId,
	})

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("text/plain")
	ctx.Response.SetBody([]byte(responseBodyOk))

	return nil
}
