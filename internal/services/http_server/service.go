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
	Port    int

	handlers *handlers.Handlers
	l        *slog.Logger
}

func New(cfg *Config, handlers *handlers.Handlers, l *slog.Logger) *HttpServer {
	return &HttpServer{
		Address:  cfg.Address,
		Port:     cfg.Port,
		handlers: handlers,
		l:        l.With("name", "HttpServer"),
	}
}

func (s *HttpServer) Run(ctx context.Context) error {
	r := router.New()
	r.GET("/metrics", s.handlers.Metrics.Handle)
	api := r.Group("/api")
	api.POST(fmt.Sprintf("/vk/callback/{%s}", vk_callback_handler.HookIdKey), s.handlers.VkCallbackHandler.Handle)

	socketAddress := fmt.Sprintf("%s:%d", s.Address, s.Port)
	l, err := net.Listen("tcp", socketAddress)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		s.l.Info("stopping http server")
		_ = l.Close()
	}()

	s.l.Info("starting http server", "address", socketAddress)
	return fasthttp.Serve(l, r.Handler)
}
