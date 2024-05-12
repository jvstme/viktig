package http_server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"viktig/internal/entities"
	"viktig/internal/queue"

	"github.com/fasthttp/router"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

const hookIdKey = "community_hook_id"

type HttpServer struct {
	Address            string
	Port               int
	MetricsAuthToken   string
	ConfirmationString string
	q                  *queue.Queue[entities.Message]
}

func New(cfg *Config, q *queue.Queue[entities.Message]) *HttpServer {
	return &HttpServer{
		Address:            cfg.Address,
		Port:               cfg.Port,
		MetricsAuthToken:   cfg.MetricsAuthToken,
		ConfirmationString: cfg.ConfirmationString,
		q:                  q,
	}
}

func (s *HttpServer) Run(ctx context.Context) error {
	r := router.New()
	r.GET(
		"/metrics",
		bearerTokenAuth(
			s.MetricsAuthToken,
			fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()),
		),
	)
	api := r.Group("/api")
	api.POST(fmt.Sprintf("/vk/callback/{%s}", hookIdKey), s.vkHandler)

	socketAddress := fmt.Sprintf("%s:%d", s.Address, s.Port)
	l, err := net.Listen("tcp", socketAddress)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		slog.Info("stopping http server")
		_ = l.Close()
	}()

	slog.Info("starting http server", "address", socketAddress)
	return fasthttp.Serve(l, r.Handler)
}
