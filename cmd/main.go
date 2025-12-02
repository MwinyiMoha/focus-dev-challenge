package main

import (
	"focus-dev-challenge/internal/adapters/repository"
	"focus-dev-challenge/internal/config"
	"log"

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

	v := validator.New()
	cfg, err := config.New(v)
	if err != nil {
		logger.Fatal("could not load config", zap.Error(err))
	}

	_, err = repository.NewRepository(cfg)
	if err != nil {
		logger.Fatal("could not initialize data repository", zap.Error(err))
	}
}
