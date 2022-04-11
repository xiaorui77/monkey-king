package api

import "github.com/xiaorui77/monker-king/internal/engine/task"

type TaskManage interface {
	DeleteTask(domain string, id uint64) *task.Task
	GetTree(domain string) interface{}
}
