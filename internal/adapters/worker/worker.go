package worker

import (
	"context"
	"focus-dev-challenge/internal/config"
	"focus-dev-challenge/internal/core/domain"
	"focus-dev-challenge/internal/core/ports"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type TaskProcessor struct {
	server     *asynq.Server
	logger     *zap.Logger
	repository ports.AppRepository
}

func NewTaskProcessor(cfg *config.Config, repo ports.AppRepository, logger *zap.Logger) *TaskProcessor {
	server := asynq.NewServer(
		&asynq.RedisClientOpt{
			Addr:        cfg.RedisHost,
			DB:          cfg.RedisDB,
			DialTimeout: time.Duration(cfg.DefaultTimeout) * time.Second,
		},
		asynq.Config{
			Queues: map[string]int{
				cfg.DefaultQueue: 10,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.Error(
					"Failed to process task",
					zap.String("original_error", err.Error()),
					zap.String("task_type", task.Type()),
					zap.ByteString("task_payload", task.Payload()),
				)
			}),
		},
	)
	return &TaskProcessor{
		server:     server,
		logger:     logger,
		repository: repo,
	}
}

func (tp *TaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(domain.SendMessageTask, tp.SendMessage)

	return tp.server.Start(mux)
}

func (tp *TaskProcessor) Stop() {
	tp.server.Shutdown()
}
