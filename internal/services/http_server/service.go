package http_server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"viktig/internal/services/http_server/handlers"
	"viktig/internal/services/http_server/handlers/vk_callback_handler"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type HttpServer struct {
	Address string

	handlers *handlers.Handlers
	l        *slog.Logger
}

func New(address string, handlers *handlers.Handlers, l *slog.Logger) *HttpServer {
	return &HttpServer{
		Address:  address,
		handlers: handlers,
		l:        l.With("name", "HttpServer"),
	}
}

func (s *HttpServer) Run(ctx context.Context) error {
	l, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		s.l.Info("stopping http server")
		_ = l.Close()
	}()

	s.l.Info("starting http server", "address", s.Address)
	return fasthttp.Serve(l, s.setupRouter().Handler)
}

func (s *HttpServer) setupRouter() *router.Router {
	r := router.New()
	if s.handlers.Metrics != nil {
		r.GET("/metrics", s.handlers.Metrics.Handle)
	}
	if s.handlers.Debug != nil {
		r.POST("/debug", s.handlers.Debug.Handle)
	}
	api := r.Group("/api")
	if s.handlers.VkCallbackHandler != nil {
		api.POST(fmt.Sprintf("/vk/callback/{%s}", vk_callback_handler.InteractionIdKey), s.handlers.VkCallbackHandler.Handle)
	}
	return r
}
