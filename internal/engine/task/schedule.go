package task

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type crawlerBrowser struct {
	priorityTask chan Task
	tasks        chan Task
	successNum   int

	cookiejar     http.CookieJar
	DefaultClient *http.Client
}

func NewRunner() Runner {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil
	}
	return &crawlerBrowser{
		priorityTask: make(chan Task, 100),
		tasks:        make(chan Task, 1000),

		cookiejar: jar,
		DefaultClient: &http.Client{
			Jar:     jar,
			Timeout: time.Second * 15,
		},
	}
}

func (r *crawlerBrowser) Run(ctx context.Context) {
	logrus.Infof("[task] Runner beginning")
	for {
		select {
		case task := <-r.priorityTask:
			r.process(ctx, r.DefaultClient, task)
		default:
			select {
			case task := <-r.tasks:
				r.process(ctx, r.DefaultClient, task)
			default:
				time.Sleep(time.Second * 10)
			}
		}
		logrus.Debugf("[task] Done, total success: %d", r.successNum)
	}
}

func (r *crawlerBrowser) process(ctx context.Context, client *http.Client, task Task) {
	switch task.(type) {
	case *downloader:
		t := task.(*downloader)
		logrus.Infof("[task] current[%d] process task: %+v", r.successNum, t.fileName)
		if err := task.Run(ctx, client); err != nil {
			time.Sleep(time.Second * 2)
			r.priorityTask <- task
			return
		}
	default:
		logrus.Infof("[task] current[%d] process task: %+v", r.successNum, task)
		if err := task.Run(ctx, client); err != nil {
			time.Sleep(time.Second * 2)
			r.tasks <- task
			return
		}
	}
	r.successNum++
}

func (r *crawlerBrowser) AddTask(t Task) {
	if task, ok := t.(*downloader); ok {
		r.priorityTask <- task
	} else if task, ok := t.(*grabber); ok {
		r.tasks <- task
		if len(r.tasks) > cap(r.tasks)-100 {
			logrus.Warnf("[task] tasks chan undercapacity: [%v/%v]", len(r.tasks), cap(r.tasks))
		}
	} else {
		r.tasks <- t
	}
}
