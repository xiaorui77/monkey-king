package task

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/yougtao/monker-king/internal/utils/math"
	"github.com/yougtao/monker-king/internal/utils/wait"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const (
	Parallelism = 2
)

type taskGrads struct {
	priority chan *Task
	normal   chan *Task
}

// Runner 是一个运行器
type Runner interface {
	Run(ctx context.Context)
	AddTask(t *Task, priority bool)
}

type crawlerBrowser struct {
	// default client
	client    *http.Client
	cookiejar http.CookieJar
	ctx       context.Context

	// 以hostname分开的队列
	queue map[string]taskGrads
}

func NewRunner() Runner {
	jar, err := cookiejar.New(nil)
	if err != nil {
		logrus.Errorf("new cookiejar failed: %v", err)
		return nil
	}
	return &crawlerBrowser{
		cookiejar: jar,
		client: &http.Client{
			Jar:     jar,
			Timeout: time.Second * 15,
		},

		queue: map[string]taskGrads{},
	}
}

func (r *crawlerBrowser) Run(ctx context.Context) {
	r.ctx = ctx
	// todo: 暂时先阻塞住
	wait.WaitWhen(func() bool { return false })
}

func (r *crawlerBrowser) AddTask(t *Task, priority bool) {
	if t == nil {
		return
	}

	if t.ID == 0 {
		t.ID = rand.Uint64()
	}

	host := t.url.Host
	if _, ok := r.queue[host]; !ok {
		r.queue[host] = taskGrads{
			priority: make(chan *Task, 100),
			normal:   make(chan *Task, 100),
		}
		go r.processHost(host)
	}

	if priority {
		r.queue[host].priority <- t
	} else {
		r.queue[host].normal <- t
	}
}

// 开启一个新hostname(url)的处理过程
func (r *crawlerBrowser) processHost(host string) {
	for i := 0; i < Parallelism; i++ {
		wait.WaitWhen(func() bool { return r.ctx != nil })
		go r.process(host)
	}
}

func (r *crawlerBrowser) process(host string) {
	last := time.Now()
	for {
		select {
		case <-r.ctx.Done():
			logrus.Infof("[task] The Runner will be cancel")
			// todo: 持久化 before return
			return
		default:
			select {
			case task := <-r.queue[host].priority:
				last = time.Now()
				logrus.Infof("[task] The task[%x] begin to run, url: %s", task.ID, task.url)
				if err := task.Run(r.ctx, r.client); err != nil {
					logrus.Errorf("[task] The task[%x] run failed(try again after): %v", task.ID, err)
					r.queue[host].priority <- task
					continue
				}
				logrus.Infof("[task] The task[%x] done.", task.ID)
			default:
				select {
				case task := <-r.queue[host].normal:
					last = time.Now()
					logrus.Infof("[task] The task[%x] begin to run", task.ID)
					if err := task.Run(r.ctx, r.client); err != nil {
						logrus.Errorf("[task] The task[%x] run failed(try again after): %v", task.ID, err)
						r.queue[host].normal <- task
						continue
					}
					logrus.Infof("[task] The task[%x] done.", task.ID)
					time.Sleep(time.Second * 10)
				default:
					sub := time.Now().Sub(last)
					if sub > time.Second*10 {
						// todo: 发送建议停止信号
					}
					time.Sleep(math.MinDuration(time.Second+sub/2, time.Second*10))
				}
			}
		}
	}
}
