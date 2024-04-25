package forwarder

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log/slog"
	"time"
)

type Forwarder struct {
	tgToken string
}

func New(cfg *Config) *Forwarder {
	return &Forwarder{
		tgToken: cfg.TgConfig.Token,
	}
}

func (f *Forwarder) Run(ctx context.Context) error {
	// пока тут просто что-то крутится
	t := time.NewTicker(time.Second)
	i := 0
	for {
		select {
		case <-t.C:
			i += 1
			slog.Info(fmt.Sprintf("%v", i))
		case <-ctx.Done():
			slog.Info("stopping...")
			return errors.New("error 1")
		}
	}
}
