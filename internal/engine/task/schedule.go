package task

import (
	"context"
	"github.com/sirupsen/logrus"
	"time"
)

// Runner 是一个运行器
type Runner interface {
	Run(ctx context.Context)
	AddTask(t Task)
}

type defaultRunner struct {
	priorityTask chan Task
	tasks        chan Task
	successNum   int
}

func NewRunner() Runner {
	return &defaultRunner{
		priorityTask: make(chan Task, 100),
		tasks:        make(chan Task, 1000),
	}
}

func (r *defaultRunner) Run(ctx context.Context) {
	logrus.Infof("[task] Runner beginning")
	for {
		select {
		case task := <-r.priorityTask:
			r.process(ctx, task)
		default:
			select {
			case task := <-r.tasks:
				r.process(ctx, task)
			default:
				time.Sleep(time.Second * 10)
			}
		}
		logrus.Debugf("[task] Done, total success: %d", r.successNum)
	}
}

func (r *defaultRunner) process(ctx context.Context, task Task) {
	switch task.(type) {
	case *downloader:
		t := task.(*downloader)
		logrus.Infof("[task] current[%d] process task: %+v", r.successNum, t.fileName)
		if err := task.Run(ctx); err != nil {
			time.Sleep(time.Second * 2)
			r.tasks <- task
			return
		}
	default:
		logrus.Infof("[task] current[%d] process task: %+v", r.successNum, task)
		if err := task.Run(ctx); err != nil {
			time.Sleep(time.Second * 2)
			r.tasks <- task
			return
		}
	}
	r.successNum++
}

func (r *defaultRunner) AddTask(t Task) {
	if task, ok := t.(*downloader); ok {
		r.priorityTask <- task
	} else if task, ok := t.(*parser); ok {
		r.tasks <- task
		if len(r.tasks) > cap(r.tasks)-100 {
			logrus.Warnf("[task] tasks chan undercapacity: [%v/%v]", len(r.tasks), cap(r.tasks))
		}
	} else {
		r.tasks <- t
	}
}
