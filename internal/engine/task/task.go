package task

import (
	"context"
	"net/http"
)

type Task interface {
	Run(ctx context.Context, client *http.Client) error
}

// Runner 是一个运行器
type Runner interface {
	Run(ctx context.Context)
	AddTask(t Task)
}
