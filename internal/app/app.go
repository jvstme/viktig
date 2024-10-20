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
	"viktig/internal/services/forwarder"
	"viktig/internal/services/http_server"
	"viktig/internal/services/vk_users_getter"

	"github.com/cosiner/flag"
	"github.com/xlab/closer"
)

type Params struct {
	ConfigPath string `names:"--config" usage:"config file path" default:"./config.yml"`
	Host       string `names:"--host" usage:"host to bind to" default:"127.0.0.1"`
	Port       int    `names:"--port" usage:"port to bind to" default:"1337"`
}

type App struct {
	params *Params
	cfg    *config.Config
}

func New() (*App, error) {
	params := &Params{}
	err := flag.Commandline.ParseStruct(params)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(params.ConfigPath)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() || (filepath.Ext(stat.Name()) != ".yaml" && filepath.Ext(stat.Name()) != ".yml") {
		return nil, fmt.Errorf("invalid config path: %s", params.ConfigPath)
	}
	cfg, err := config.LoadConfigFromFile(params.ConfigPath)
	if err != nil {
		return nil, err
	}
	return &App{params, cfg}, nil
}

func (a App) Run() {
	errorCh := make(chan error)
	appCtx, wg := setupContextAndWg(context.Background(), errorCh)

	q1 := queue.NewQueue[entities.Message]() // callback_handler --> users_getter
	q2 := queue.NewQueue[entities.Message]() // users_getter --> forwarder

	vkUsersGetterService := vk_users_getter.New(a.cfg.VkApiToken, q1, q2, slog.Default())
	wg.Add(1)
	go func() {
		defer wg.Done()
		errorCh <- vkUsersGetterService.Run(appCtx)
	}()

	forwarderService := a.makeForwarder(q2)
	wg.Add(1)
	go func() {
		defer wg.Done()
		errorCh <- forwarderService.Run(appCtx)
	}()

	httpServer := a.makeHttpServer(q1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		errorCh <- httpServer.Run(appCtx)
	}()

	closer.Hold()
}

func (a App) makeHttpServer(q *queue.Queue[entities.Message]) *http_server.HttpServer {
	communities := make(map[string]*http_server.Community)
	for _, community := range a.cfg.Communities {
		communities[community.HookId] = &http_server.Community{
			SecretKey:          community.SecretKey,
			ConfirmationString: community.ConfirmationString,
		}
	}
	return http_server.New(
		a.params.Host,
		a.params.Port,
		a.cfg.MetricsAuthToken,
		communities,
		q,
		slog.Default(),
	)
}

func (a App) makeForwarder(q *queue.Queue[entities.Message]) *forwarder.Forwarder {
	communities := make(map[string]*forwarder.Community)
	for _, community := range a.cfg.Communities {
		communities[community.HookId] = &forwarder.Community{TgChatId: community.TgChatId}
	}
	return forwarder.New(a.cfg.TgBotToken, communities, q, slog.Default())
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
