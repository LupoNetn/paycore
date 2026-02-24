package tasks

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/hibiken/asynq"
)

func HandleSendOTPEmailTask(ctx context.Context, t *asynq.Task) error {
	var payload SendOTPEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		slog.Error("failed to unmarshal send otp email payload", "error", err)
		return err
	}
	return nil
}
