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

type Community struct {
	SecretKey          string
	ConfirmationString string
}

type HttpServer struct {
	host             string
	port             int
	metricsAuthToken string
	communities      map[string]*Community
	q                *queue.Queue[entities.Message]
	l                *slog.Logger
}

func New(
	host string,
	port int,
	metricsAuthToken string,
	communities map[string]*Community,
	q *queue.Queue[entities.Message],
	l *slog.Logger,
) *HttpServer {
	return &HttpServer{
		host:             host,
		port:             port,
		metricsAuthToken: metricsAuthToken,
		communities:      communities,
		q:                q,
		l:                l.With("service", "HttpServer"),
	}
}

func (s *HttpServer) Run(ctx context.Context) error {
	r := router.New()
	if s.metricsAuthToken != "" {
		r.GET(
			"/metrics",
			bearerTokenAuth(
				s.metricsAuthToken,
				fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()),
			),
		)
	}
	api := r.Group("/api")
	api.POST(fmt.Sprintf("/vk/callback/{%s}", hookIdKey), s.vkHandler)

	socketAddress := fmt.Sprintf("%s:%d", s.host, s.port)
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
