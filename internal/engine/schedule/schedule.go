package schedule

import (
	"context"
	"github.com/xiaorui77/goutils/wait"
	"github.com/xiaorui77/monker-king/internal/engine/download"
	"github.com/xiaorui77/monker-king/internal/engine/task"
	"github.com/xiaorui77/monker-king/internal/storage"
	"github.com/xiaorui77/monker-king/pkg/model"
	"sort"
	"time"
)

const (
	// Parallelism is maximum concurrent number of the same domain.
	Parallelism = 3

	taskQueueSize = 100
)

type Scheduler struct {
	ctx      context.Context
	download *download.Downloader
	store    storage.Store

	taskQueue chan *task.Task
	// 以domain分开的队列
	browsers map[string]*DomainBrowser
}

func NewRunner(store storage.Store) *Scheduler {
	s := &Scheduler{
		taskQueue: make(chan *task.Task, taskQueueSize),
		browsers:  map[string]*DomainBrowser{},
		store:     store,
	}
	s.initIdentify()
	return s
}

// Run in Blocking mode
func (s *Scheduler) Run(ctx context.Context) {
	s.ctx = ctx
	s.download = download.NewDownloader(ctx)

	for {
		select {
		case <-ctx.Done():
			s.close()
			wait.WaitUntil(func() bool { return len(s.browsers) == 0 })
			return
		case t := <-s.taskQueue:
			t.SetState(task.StateInit)
			host := s.obtainDomain(t.Url)
			if _, ok := s.browsers[host]; !ok {
				s.browsers[host] = NewDomainBrowser(s, host)
				go s.browsers[host].begin(ctx)
			}
			s.browsers[host].push(t)
		}
	}
}

func (s *Scheduler) AddTask(t *task.Task) {
	if t != nil {
		s.taskQueue <- t
	}
}

func (s *Scheduler) GetRows() []interface{} {
	now := time.Now()
	rows := make([]interface{}, 0, len(s.browsers))

	for _, domain := range s.browsers {
		ls := domain.list()
		// 默认排序: state,time
		sort.SliceStable(ls, func(i, j int) bool {
			if ls[i].State == ls[j].State {
				return ls[i].Time.Unix() > ls[j].Time.Unix()
			}
			return ls[i].State < ls[j].State
		})

		for _, t := range ls {
			rows = append(rows, &model.TaskRow{
				ID:     t.ID,
				Name:   t.Name,
				Domain: domain.domain,
				State:  task.StateStatus[t.State],
				URL:    t.Url.String(),
				Age:    now.Sub(t.Time).Truncate(time.Second).String(),
			})
		}
	}
	return rows
}

func (s *Scheduler) close() {
	// todo: 保存状态
}
