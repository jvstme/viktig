package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"viktig/internal/config"
	"viktig/internal/entities"
	"viktig/internal/queue"
	"viktig/internal/repository"
	"viktig/internal/repository/sqlite_repo"
	"viktig/internal/services/http_server"
	"viktig/internal/services/http_server/handlers"
	"viktig/internal/services/http_server/handlers/debug_handler"
	"viktig/internal/services/http_server/handlers/metrics_handler"
	"viktig/internal/services/http_server/handlers/vk_callback_handler"

	"github.com/xlab/closer"
)

type App struct {
	configPath string
}

func New() App {
	path := "./configs/example.yaml"
	if len(os.Args) == 2 {
		path = os.Args[1]
	}
	stat, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	if stat.IsDir() || (filepath.Ext(stat.Name()) != ".yaml" && filepath.Ext(stat.Name()) != ".yml") {
		panic(fmt.Sprintf("invalid config path: %s", path))
	}
	return App{configPath: path}
}

func (a App) Run() error {
	cfg, err := config.LoadConfigFromFile(a.configPath)
	if err != nil {
		return err
	}

	errorCh := make(chan error)
	appCtx, wg := setupContextAndWg(context.Background(), errorCh)

	q := queue.NewQueue[entities.Message]()
	repo := sqlite_repo.New()

	//forwarderService := forwarder.New(cfg.ForwarderConfig, q, repo, slog.Default())
	//wg.Add(1)
	//go func() {
	//	defer wg.Done()
	//	errorCh <- forwarderService.Run(appCtx)
	//}()

	httpServer := http_server.New(cfg.HttpServerConfig, setupHandlers(cfg, q, repo), slog.Default())
	wg.Add(1)
	go func() {
		defer wg.Done()
		errorCh <- httpServer.Run(appCtx)
	}()

	// Предлагаю 3 сервиса:
	// 		[x] Сервис 1: веб-сервер на который хукается вк. Он кладет сообщение во внешнюю очередь
	// 		[x] Сервис 2: шлет сообщения из очереди в нужные каналы. Можно добавить ретраи
	// 		[x] Очередь
	//      [x] К ним репо для БД. В бд храним инфу о пользователе и взаимодействиях иже с ней
	// todo [ ] Сервис 3: UI бота/настройки

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

func setupHandlers(cfg *config.Config, q *queue.Queue[entities.Message], repo repository.Repository) *handlers.Handlers {
	var debug handlers.Handler
	if os.Getenv("RUN_ENV") == "DEBUG" {
		debug = debug_handler.New(cfg.HttpServerConfig.Host, repo)
	}
	return &handlers.Handlers{
		VkCallbackHandler: vk_callback_handler.New(q, repo, slog.Default()),
		Metrics:           metrics_handler.New(cfg.MetricsConfig),
		Debug:             debug,
		//todo: TgBotUIHandler?
	}
}
