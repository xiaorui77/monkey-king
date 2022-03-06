package schedule

import (
	"context"
	"github.com/yougtao/goutils/logx"
	"github.com/yougtao/goutils/wait"
	"github.com/yougtao/monker-king/internal/storage"
	"github.com/yougtao/monker-king/internal/view/model"
	"net/http"
	"net/http/cookiejar"
	"sort"
	"time"
)

const (
	// Parallelism is maximum concurrent number of the same host
	Parallelism = 3
)

// Runner 是一个运行器
type Runner interface {
	Run(ctx context.Context)
	AddTask(t *Task, priority bool)
}

type Scheduler struct {
	// default client
	client    *http.Client
	cookiejar http.CookieJar
	ctx       context.Context

	// 以hostname分开的队列
	queue map[string]*DomainBrowser
	store storage.Store
}

func NewRunner(store storage.Store) *Scheduler {
	jar, err := cookiejar.New(nil)
	if err != nil {
		logx.Errorf("new cookiejar failed: %v", err)
		return nil
	}
	return &Scheduler{
		cookiejar: jar,
		client: &http.Client{
			Jar:     jar,
			Timeout: time.Second * 15,
		},

		queue: map[string]*DomainBrowser{},
		store: store,
	}
}

func (r *Scheduler) Run(ctx context.Context) {
	r.ctx = ctx

	<-ctx.Done()
	wait.WaitUntil(func() bool { return len(r.queue) == 0 })
}

func (r *Scheduler) AddTask(t *Task, priority bool) {
	if t == nil {
		return
	}

	host := t.url.Host
	if _, ok := r.queue[host]; !ok {
		r.queue[host] = NewDomainBrowser(host)
		go r.queue[host].Schedule(r.ctx)
	}

	r.queue[host].Push(priority, t)
}

func (r *Scheduler) GetRows() []interface{} {
	now := time.Now()
	rows := make([]interface{}, 0, len(r.queue))

	for _, domain := range r.queue {
		ls := domain.List()
		// 默认排序: state,time
		sort.Slice(ls, func(i, j int) bool {
			if ls[i].state == ls[j].state {
				return ls[i].time.Unix() < ls[j].time.Unix()
			}
			return ls[i].state < ls[j].state
		})

		for _, t := range ls {
			rows = append(rows, &model.TaskRow{
				ID:     t.ID,
				Name:   t.Name,
				Domain: domain.domain,
				State:  TaskStateStatus[t.state],
				URL:    t.url.String(),
				Age:    now.Sub(t.time).Truncate(time.Second).String(),
			})
		}
	}
	return rows
}
