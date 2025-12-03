package worker

import (
	"context"
	"encoding/json"
	"focus-dev-challenge/internal/core/domain"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func (tp *TaskProcessor) SendMessage(ctx context.Context, task *asynq.Task) error {
	var payload domain.SendMessage
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		tp.logger.Error("failed to parse task data", zap.Error(err))
		return err
	}

	tp.logger.Info("sending message", zap.Int64("message_id", payload.MessageID))

	// Message sending logic goes here

	return nil
}
