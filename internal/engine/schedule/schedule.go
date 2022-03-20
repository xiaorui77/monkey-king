package schedule

import (
	"context"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/goutils/wait"
	"github.com/xiaorui77/monker-king/internal/storage"
	"github.com/xiaorui77/monker-king/pkg/model"
	"net"
	"net/http"
	"net/http/cookiejar"
	"sort"
	"time"
)

const (
	// Parallelism is maximum concurrent number of the same host
	Parallelism = 5
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
	browsers map[string]*DomainBrowser
	store    storage.Store
}

func NewRunner(store storage.Store) *Scheduler {
	jar, err := cookiejar.New(nil)
	if err != nil {
		logx.Fatalf("[scheduler] new cookiejar failed: %v", err)
		return nil
	}
	s := &Scheduler{
		cookiejar: jar,
		client: &http.Client{
			Jar:     jar,
			Timeout: time.Second * 15,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   15 * time.Second,
					KeepAlive: 10 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   10,
				IdleConnTimeout:       60 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},

		browsers: map[string]*DomainBrowser{},
		store:    store,
	}
	s.initIdentify()
	return s
}

// Run in Blocking mode
func (s *Scheduler) Run(ctx context.Context) {
	s.ctx = ctx

	<-ctx.Done()
	wait.WaitUntil(func() bool { return len(s.browsers) == 0 })
}

func (s *Scheduler) AddTask(t *Task, priority bool) {
	if t == nil {
		return
	}
	t.state = TaskStateInit
	host := s.obtainDomain(t.url)

	if _, ok := s.browsers[host]; !ok {
		s.browsers[host] = NewDomainBrowser(host)
		go s.browsers[host].begin(s.ctx)
	}
	s.browsers[host].push(priority, t)
}

func (s *Scheduler) GetRows() []interface{} {
	now := time.Now()
	rows := make([]interface{}, 0, len(s.browsers))

	for _, domain := range s.browsers {
		ls := domain.list()
		// 默认排序: state,time
		sort.SliceStable(ls, func(i, j int) bool {
			if ls[i].state == ls[j].state {
				return ls[i].time.Unix() > ls[j].time.Unix()
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
