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
	"viktig/internal/services/forwarder"
	"viktig/internal/services/http_server"
	"viktig/internal/services/http_server/handlers"
	"viktig/internal/services/http_server/handlers/debug_handler"
	"viktig/internal/services/http_server/handlers/metrics_handler"
	"viktig/internal/services/http_server/handlers/vk_callback_handler"

	"github.com/cosiner/flag"
	"github.com/xlab/closer"
)

type Params struct {
	Debug      bool   `names:"--debug" usage:"enable /debug route" default:"false"`
	ConfigPath string `names:"--config, -c" usage:"config file path" default:"./config.yaml"`
	Address    string `names:"--address, -a" usage:"address of machine or container" default:"127.0.0.1:8080"`
	Host       string `names:"--host" usage:"domain name for requests" default:"example.com"`
}

type App struct {
	params *Params
}

func New() App {
	return App{params: getValidatedParams()}
}

func (a App) Run() error {
	cfg, err := config.LoadConfigFromFile(a.params.ConfigPath)
	if err != nil {
		return err
	}

	errorCh := make(chan error)
	appCtx, wg := a.setupContextAndWg(context.Background(), errorCh)

	q := queue.NewQueue[entities.Message]()
	repo := sqlite_repo.New(cfg.RepoConfig)

	forwarderService := forwarder.New(cfg.ForwarderConfig, q, repo, slog.Default())
	wg.Add(1)
	go func() {
		defer wg.Done()
		errorCh <- forwarderService.Run(appCtx)
	}()

	httpServer := http_server.New(a.params.Address, a.setupHandlers(cfg, q, repo), slog.Default())
	wg.Add(1)
	go func() {
		defer wg.Done()
		errorCh <- httpServer.Run(appCtx)
	}()

	closer.Hold()
	return nil
}

// setupContextAndWg returns a context cancelled on app shutdown request and a wait group awaited on shutdown.
//
//	All non-nil errors received from errorCh after an app shutdown request will be logged as "App shutdown errors".
//	If an error is received from errorCh before an app shutdown request, closer.Close will be called.
func (a App) setupContextAndWg(parentCtx context.Context, errorCh chan error) (ctx context.Context, wg *sync.WaitGroup) {
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

func (a App) setupHandlers(cfg *config.Config, q *queue.Queue[entities.Message], repo repository.Repository) *handlers.Handlers {
	var debug handlers.Handler
	if a.params.Debug {
		debug = debug_handler.New(a.params.Host, repo)
	}
	return &handlers.Handlers{
		VkCallbackHandler: vk_callback_handler.New(q, repo, slog.Default()),
		Metrics:           metrics_handler.New(cfg.MetricsConfig),
		Debug:             debug,
		//todo: TgBotUIHandler?
	}
}

func getValidatedParams() *Params {
	params := &Params{}
	err := flag.Commandline.ParseStruct(params)
	if err != nil {
		panic(err)
	}

	stat, err := os.Stat(params.ConfigPath)
	if err != nil {
		panic(err)
	}
	if stat.IsDir() || (filepath.Ext(stat.Name()) != ".yaml" && filepath.Ext(stat.Name()) != ".yml") {
		panic(fmt.Sprintf("invalid config path: %s", params.ConfigPath))
	}

	return params
}
