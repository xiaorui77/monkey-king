package types

import (
	"github.com/xiaorui77/monker-king/internal/engine/schedule"
	"github.com/xiaorui77/monker-king/internal/engine/task"
	"github.com/xiaorui77/monker-king/internal/view/model"
)

type Collect interface {
	Visit(parent *task.Task, url string) error

	Scheduler() *schedule.Scheduler
	GetDataProducer() model.DataProducer
}
