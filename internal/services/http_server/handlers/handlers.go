package handlers

import (
	"github.com/valyala/fasthttp"
)

type Handler interface {
	Handle(ctx *fasthttp.RequestCtx)
}

type Handlers struct {
	VkCallbackHandler Handler
	Metrics           Handler
	Debug             Handler
	TelegramBot       Handler
}
