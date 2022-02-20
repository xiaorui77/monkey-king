package task

import (
	"context"
	"encoding/json"
	"github.com/yougtao/goutils/logx"
	"github.com/yougtao/goutils/math"
	"github.com/yougtao/goutils/wait"
	"github.com/yougtao/monker-king/internal/storage"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"
)

const (
	Parallelism = 2
)

type taskGrads struct {
	priority chan *Task
	normal   chan *Task
}

func (s taskGrads) MarshalJSON() ([]byte, error) {
	logx.Debugf("MarshalJson: %v", s)
	data := map[string][]Task{}
	if len(s.priority) != 0 {
		logx.Debugf("MarshalJson for from chan: %+v", s.priority)
		data["priority"] = closeChannelAndGet(s.priority)
		logx.Debugf("MarshalJson priority data finish：%v", data["priority"])
	}
	if len(s.normal) != 0 {
		data["normal"] = closeChannelAndGet(s.normal)
	}
	logx.Debugf("[task] chan data to map data: %v", data)
	return json.Marshal(data)
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
	store storage.Store
}

func NewRunner(store storage.Store) Runner {
	jar, err := cookiejar.New(nil)
	if err != nil {
		logx.Errorf("new cookiejar failed: %v", err)
		return nil
	}
	return &crawlerBrowser{
		cookiejar: jar,
		client: &http.Client{
			Jar:     jar,
			Timeout: time.Second * 15,
		},

		queue: map[string]taskGrads{},
		store: store,
	}
}

func (r *crawlerBrowser) Run(ctx context.Context) {
	r.ctx = ctx

	<-ctx.Done()
	wait.WaitUntil(func() bool { return len(r.queue) == 0 })
}

func (r *crawlerBrowser) AddTask(t *Task, priority bool) {
	if t == nil {
		return
	}

	if t.ID == 0 {
		t.ID = rand.Uint64()
	}

	host := t.Url.Host
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

// 开启一个新hostname(url)的处理过程, 并守护该处理过程
func (r *crawlerBrowser) processHost(host string) {
	var wg sync.WaitGroup
	for i := 0; i < Parallelism; i++ {
		wait.WaitUntil(func() bool { return r.ctx != nil })
		go r.process(&wg, host, i)
		wg.Add(1)
	}

	wg.Wait()
	// 持久化
	logx.Infof("[task] The processHost[%s] had been finished, will be persistence data", host)
	if err := r.store.PersistenceTasks(host, r.queue[host]); err != nil {
		logx.Errorf("[task] stop precessHost[%s] failed: %v", host, err)
		time.Sleep(time.Second * 30)
	}
	delete(r.queue, host)
}

func (r *crawlerBrowser) process(wg *sync.WaitGroup, host string, index int) {
	last := time.Now()
	for {
		select {
		case <-r.ctx.Done():
			logx.Infof("[task] The process-%d[%s] will be cancel", index, host)
			wg.Done()
			return
		default:
			select {
			case task := <-r.queue[host].priority:
				last = time.Now()
				logx.Infof("[task] The task[%x] begin to run, url: %s", task.ID, task.Url)
				if err := task.Run(r.ctx, r.client); err != nil {
					logx.Warnf("[task] The task[%x] run failed(try again after): %v", task.ID, err)
					r.queue[host].priority <- task
					continue
				}
				logx.Infof("[task] The task[%x] done.", task.ID)
			default:
				select {
				case task := <-r.queue[host].normal:
					last = time.Now()
					logx.Infof("[task] The task[%x] begin to run", task.ID)
					if err := task.Run(r.ctx, r.client); err != nil {
						logx.Warnf("[task] The task[%x] run failed(try again after): %v", task.ID, err)
						r.queue[host].normal <- task
						continue
					}
					logx.Infof("[task] The task[%x] done.", task.ID)
					time.Sleep(time.Second * 10)
				default:
					sub := time.Now().Sub(last)
					if sub > time.Second*15 {
						// todo: 发送建议停止信号
					}
					time.Sleep(math.MinDuration(time.Second+sub/2, time.Second*15))
				}
			}
		}
	}
}

func closeChannelAndGet(ch chan *Task) []Task {
	var tasks = make([]Task, 0, len(ch))
	close(ch)
	for t := range ch {
		tasks = append(tasks, *t)
	}
	return tasks
}
