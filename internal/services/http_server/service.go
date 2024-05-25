package http_server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"viktig/internal/services/http_server/handlers"
	"viktig/internal/services/http_server/handlers/vk_callback_handler"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

const defaultPort = 8080

type HttpServer struct {
	Address string
	Port    int

	handlers *handlers.Handlers
	l        *slog.Logger
}

func New(address, port string, handlers *handlers.Handlers, l *slog.Logger) *HttpServer {
	intPort, err := strconv.Atoi(port)
	if err != nil {
		intPort = defaultPort
	}
	return &HttpServer{
		Address:  address,
		Port:     intPort,
		handlers: handlers,
		l:        l.With("name", "HttpServer"),
	}
}

func (s *HttpServer) Run(ctx context.Context) error {
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
