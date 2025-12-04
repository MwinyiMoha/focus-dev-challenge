package main

import (
	"context"
	"fmt"
	"focus-dev-challenge/internal/adapters/api"
	"focus-dev-challenge/internal/adapters/repository"
	"focus-dev-challenge/internal/adapters/worker"
	"focus-dev-challenge/internal/config"
	"focus-dev-challenge/internal/core/app"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/mwinyimoha/commons/pkg/logging"
	"go.uber.org/zap"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	logger, err := logging.NewLoggerConfig().BuildLogger()
	if err != nil {
		log.Fatal("could not initialize logger")
	}

	defer func() { _ = logger.Sync() }()

	val := validator.New()
	cfg, err := config.New(val)
	if err != nil {
		logger.Fatal("could not load config", zap.Error(err))
	}

	repo, err := repository.NewRepository(cfg)
	if err != nil {
		logger.Fatal("could not initialize data repository", zap.Error(err))
	}

	svc := app.NewService(cfg, repo, val)
	ch := make(chan error, 1)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	var srv *http.Server
	var tasker *worker.TaskProcessor

	switch cfg.AppTier {
	case "web":
		router := api.NewRouter(svc, logger, cfg.Debug)
		srv = &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
			Handler: router.Engine,
		}

		go func() {
			logger.Info(
				"starting server",
				zap.String("app_name", cfg.ServiceName),
				zap.Int("port", cfg.ServerPort),
			)

			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("failed to start server", zap.Error(err))
				ch <- err
			}
		}()
	case "worker":
		tasker = worker.NewTaskProcessor(cfg, repo, logger)

		go func() {
			logger.Info("Starting background task processor")

			if err := tasker.Start(); err != nil {
				logger.Error("could not start task processor", zap.Error(err))
				ch <- err
			}
		}()
	default:
		logger.Fatal("unsupported app tier", zap.String("tier", cfg.AppTier))
	}

	select {
	case <-ctx.Done():
		logger.Info("initiating graceful shutdown")

		if cfg.AppTier == "web" {
			if err := srv.Shutdown(context.Background()); err != nil {
				logger.Error("failed graceful shutdown", zap.Error(err))
			}
		} else {
			tasker.Stop()
		}

		if err := repo.Close(); err != nil {
			logger.Error("failed to close repository", zap.Error(err))
		}

		logger.Info("application stopped")
	case err := <-ch:
		logger.Fatal("application stopped with error", zap.Error(err))
	}
}
