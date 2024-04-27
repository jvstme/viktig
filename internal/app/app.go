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

	appCtx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	errorCh := make(chan error)

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

	// all other services go here
	// all services must shut down on <-appCtx.Done() and return an error

	// Предлагаю 3 сервиса:
	// 		[x] Сервис 1: веб-сервер на который хукается вк. Он кладет сообщение во внешнюю очередь
	// 		[x] Сервис 2: шлет сообщения из очереди в нужные каналы. Можно добавить ретраи
	// 		[x] Очередь
	// todo [ ] Сервис 3: UI бота/настройки
	// todo [ ] К ним репо для БД. В бд храним инфу о пользователе и иже с ней

	closer.Bind(gatherErrors(errorCh))
	closer.Bind(func() { close(errorCh) })
	closer.Bind(wg.Wait)
	closer.Bind(cancel)
	closer.Hold()

	return nil
}

func gatherErrors(errorCh <-chan error) func() {
	resCh := make(chan error)
	go func() {
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
		resCh <- res
	}()

	return func() {
		slog.Error(fmt.Sprintf("app shutdown errors:\n%v", <-resCh))
	}
}
