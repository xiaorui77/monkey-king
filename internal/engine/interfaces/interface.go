package interfaces

import (
	"github.com/xiaorui77/monker-king/internal/engine/task"
	"github.com/xiaorui77/monker-king/internal/engine/types"
	"github.com/xiaorui77/monker-king/internal/view/model"
	error2 "github.com/xiaorui77/monker-king/pkg/error"
)

type Collect interface {
	Visit(parent *task.Task, url string) error

	TaskManager() TaskManage
	GetDataProducer() model.DataProducer
}

type TaskManage interface {
	DeleteTask(domain, name string) *task.Task
}

type Parsing interface {
	HandleOnResponse(resp *types.ResponseWarp) error2.Error
}
