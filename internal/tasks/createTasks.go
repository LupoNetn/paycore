package tasks

import (
	"encoding/json"
	"log/slog"

	"github.com/hibiken/asynq"
)

func NewSendOTPEmailTask(payload SendOTPEmailPayload) (*asynq.Task, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal send otp email payload", "error", err)
		return nil, err
	}

	return asynq.NewTask(TypeSendOTPEmail, payloadBytes), nil
}
