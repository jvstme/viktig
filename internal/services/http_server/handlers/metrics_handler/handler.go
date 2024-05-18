package metrics_handler

import (
	"viktig/internal/services/http_server"
	"viktig/internal/services/http_server/handlers"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type metricsHandler struct {
	token string
}

func New(cfg *Config) handlers.Handler {
	return &metricsHandler{token: cfg.MetricsAuthToken}
}

func (h *metricsHandler) Handle(ctx *fasthttp.RequestCtx) {
	http_server.BearerTokenAuth(h.token, fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()))(ctx)
}
