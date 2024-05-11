package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"viktig/internal/config"
	"viktig/internal/entities"
	"viktig/internal/queue"
	"viktig/internal/services/forwarder"
	"viktig/internal/services/http_server"

	"github.com/xlab/closer"
)

type App struct {
}

func New() App {
	// todo: cfg path from binary args
	return App{}
}

func (a App) Run() error {
	cfg, err := config.LoadConfigFromFile("./configs/example.yaml")
	if err != nil {
		return err
	}

	errorCh := make(chan error)
	appCtx, wg := setupContextAndWg(context.Background(), errorCh)

	q := queue.NewQueue[entities.Message]()

	forwarderService := forwarder.New(cfg.ForwarderConfig, q)
	wg.Add(1)
	go func() {
		defer wg.Done()
		errorCh <- forwarderService.Run(appCtx)
	}()

	httpServer := http_server.New(cfg.HttpServerConfig, q)
	wg.Add(1)
	go func() {
		defer wg.Done()
		errorCh <- httpServer.Run(appCtx)
	}()

	// all services go here
	// all services must shut down on <-appCtx.Done() and return an error

	// Предлагаю 3 сервиса:
	// 		[x] Сервис 1: веб-сервер на который хукается вк. Он кладет сообщение во внешнюю очередь
	// 		[x] Сервис 2: шлет сообщения из очереди в нужные каналы. Можно добавить ретраи
	// 		[x] Очередь
	// todo [ ] Сервис 3: UI бота/настройки
	// todo [ ] К ним репо для БД. В бд храним инфу о пользователе и иже с ней

	closer.Hold()
	return nil
}

// setupContextAndWg returns a context cancelled on app shutdown request and a wait group awaited on shutdown.
//
//	All non-nil errors received from errorCh after an app shutdown request will be logged as "App shutdown errors".
//	If an error is received from errorCh before an app shutdown request, closer.Close will be called.
func setupContextAndWg(parentCtx context.Context, errorCh chan error) (ctx context.Context, wg *sync.WaitGroup) {
	wg = &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(parentCtx)

	go func() {
		select {
		case <-ctx.Done():
			return
		case err := <-errorCh:
			slog.Error(fmt.Sprintf("stopping due to error: %+v", err))
			closer.Close()
		}
	}()

	closer.Bind(func() {
		var res error
		for err := range errorCh {
			if err == nil {
				continue
			}
			if res == nil {
				res = fmt.Errorf("%+v", err)
			}
			res = fmt.Errorf("%s\n%+v", res, err)
		}

		if res != nil {
			slog.Error(fmt.Sprintf("App shutdown errors:\n%+v", res))
		}
	})
	closer.Bind(func() {
		go func() {
			wg.Wait()
			close(errorCh)
		}()
	})
	closer.Bind(cancel)

	return
}
