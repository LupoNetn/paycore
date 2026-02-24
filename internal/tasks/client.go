package tasks

import (
	"github.com/hibiken/asynq"
)

func NewTaskClient(redisAddr string) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
}
