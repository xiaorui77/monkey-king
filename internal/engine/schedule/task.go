package schedule

import (
	"context"
	"fmt"
	"github.com/yougtao/goutils/logx"
	"github.com/yougtao/goutils/math"
	"github.com/yougtao/goutils/wait"
	"github.com/yougtao/monker-king/internal/utils"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Task struct {
	ID    uint64
	state int
	Url   *url.URL
	time  time.Time
	fun   callback
}

const (
	TaskStateKnown = iota
	TaskStateInit
	TaskStateRunning
	TaskStateSuccess
	TaskStateFail
)

var TaskStateStatus = map[int]string{
	1: "Init",
	2: "running",
	3: "Success",
	4: "Fail",
}

// 临时, 后面搞一下downloader
var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

func NewTask(u *url.URL, fun callback) *Task {
	return &Task{
		ID:    0,
		Url:   u,
		state: TaskStateInit,
		time:  time.Now(),
		fun:   fun,
	}
}

func (task *Task) Run(ctx context.Context, client *http.Client) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, task.Url.String(), nil)
	if err != nil {
		logx.Warnf("[schedule] The schedule[%x] failed during the new request: %v", task.ID, err)
		return fmt.Errorf("new request fail: %v", err)
	}
	req.Header.Set(utils.UserAgentKey, utils.RandomUserAgent())

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		logx.Warnf("[schedule] The schedule[%x] failed during the do request: %v", task.ID, err)
		return fmt.Errorf("do request fail")
	}

	if resp.StatusCode == http.StatusOK {
		return task.fun(req, resp)
	} else {
		logx.Warnf("[schedule] The schedule[%x] failed with unknown status code[%d]", task.ID, resp.StatusCode)
		return fmt.Errorf("do request fail with status code[%d]", resp.StatusCode)
	}
}

// 请求完成后回调
type callback func(req *http.Request, resp *http.Response) error

// TaskQueue 任务队列
type TaskQueue struct {
	sync.RWMutex

	tasks  []*Task
	index  int
	offset int
}

func NewTaskQueue(tasks []*Task) *TaskQueue {
	return &TaskQueue{
		tasks: tasks,
	}
}

func (tq *TaskQueue) Push(task *Task) {
	tq.Lock()
	tq.tasks = append(tq.tasks, task)
	tq.Unlock()
}

func (tq *TaskQueue) Next() *Task {
	tq.Lock()
	defer tq.Unlock()

	if tq.offset >= len(tq.tasks) {
		return nil
	}
	return tq.tasks[tq.offset]
}

func (tq *TaskQueue) List() []*Task {
	return tq.tasks
}

func (tq *TaskQueue) ListOption(status int) []*Task {
	tq.RLock()
	tq.RUnlock()

	res := make([]*Task, 0, len(tq.tasks)/2)
	for _, task := range tq.tasks {
		if task.state == status {
			res = append(res, task)
		}
	}
	return res
}

// --------------------------------------------------

// DomainBrowser 在同一个域名下的调度器
type DomainBrowser struct {
	domain string

	priority *TaskQueue
	normal   *TaskQueue
}

func NewHostDomain(host string) *DomainBrowser {
	return &DomainBrowser{
		domain: host,
	}
}

func (d *DomainBrowser) Push(priority bool, task *Task) {
	if priority {
		d.priority.Push(task)
	} else {
		d.normal.Push(task)
	}
}

func (d *DomainBrowser) Next() *Task {
	if task := d.priority.Next(); task != nil {
		return task
	}
	return d.normal.Next()
}

func (d *DomainBrowser) List() []*Task {
	return d.priority.List()
}

// Schedule begin schedule all tasks by multi-thread.
func (d *DomainBrowser) Schedule(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < Parallelism; i++ {
		// 因为可能在创建ctx之前, 已经有任务被添加进来了
		wait.WaitUntil(func() bool { return ctx != nil })
		go d.process(ctx, &wg, i)
		wg.Add(1)
	}

	wg.Wait()
	logx.Infof("[schedule] The Schedule of domain[%s] will be closed, but the data will be saved before that", d.domain)
	d.close()
}

func (d *DomainBrowser) process(ctx context.Context, wg *sync.WaitGroup, index int) {
	last := time.Now()
	for {
		select {
		case <-ctx.Done():
			logx.Infof("[schedule] The process-%d[%s] will stop", index, d.domain)
			wg.Done()
			return
		default:
			task := d.Next()
			if task == nil {
				sub := time.Now().Sub(last)
				if sub > time.Second*15 {
					// todo: 发送建议停止信号
				}
				time.Sleep(math.MinDuration(time.Second+sub/2, time.Second*15))
			}
			last = time.Now()
			logx.Infof("[schedule] The schedule[%x] begin to run, url: %s", task.ID, task.Url)
			if err := task.Run(ctx, httpClient); err != nil {
				logx.Warnf("[schedule] The schedule[%x] run failed(try again after): %v", task.ID, err)
				// todo: new add task
				continue
			}
			logx.Infof("[schedule] The schedule[%x] done.", task.ID)
			time.Sleep(time.Second * 10)
		}
	}
}

func (d *DomainBrowser) close() {
	// todo: close all task queue of the domain
}

func (d *DomainBrowser) MarshalJSON() ([]byte, error) {
	return nil, nil
}
