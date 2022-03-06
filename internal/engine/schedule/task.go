package schedule

import (
	"context"
	"fmt"
	"github.com/rfyiamcool/backoff"
	"github.com/yougtao/goutils/logx"
	"github.com/yougtao/goutils/wait"
	"github.com/yougtao/monker-king/internal/utils"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// 请求完成后回调
type callback func(req *http.Request, resp *http.Response) error

type Task struct {
	ID    uint64
	Name  string
	state int
	url   *url.URL
	time  time.Time
	fun   callback
}

const (
	TaskStateKnown = iota
	TaskStateRunning
	TaskStateInit
	TaskStateFail
	TaskStateSuccess
)

var TaskStateStatus = map[int]string{
	1: "running",
	2: "init",
	3: "Fail",
	4: "Success",
}

// 临时, 后面搞一下downloader
var httpClient = &http.Client{
	Timeout: time.Second * 60,
}

func NewTask(name string, u *url.URL, fun callback) *Task {
	return &Task{
		ID:    0,
		Name:  name,
		url:   u,
		state: TaskStateInit,
		time:  time.Now(),
		fun:   fun,
	}
}

func (t *Task) Run(ctx context.Context, client *http.Client) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.url.String(), nil)
	if err != nil {
		logx.Warnf("[schedule] The schedule[%x] failed during the new request: %v", t.ID, err)
		return fmt.Errorf("new request fail: %v", err)
	}
	req.Header.Set(utils.UserAgentKey, utils.RandomUserAgent())

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		logx.Warnf("[schedule] The schedule[%x] failed during the do request: %v", t.ID, err)
		return fmt.Errorf("do request fail")
	}

	if resp.StatusCode == http.StatusOK {
		return t.fun(req, resp)
	} else {
		logx.Warnf("[schedule] The schedule[%x] failed with unknown status code[%d]", t.ID, resp.StatusCode)
		return fmt.Errorf("do request fail with status code[%d]", resp.StatusCode)
	}
}

func (t *Task) SetState(state int) {
	t.state = state
}

func (t *Task) String() string {
	return fmt.Sprintf("[%s] %s", t.Name, t.url.String())
}

// TaskQueue 任务队列
type TaskQueue struct {
	sync.RWMutex

	tasks  []*Task
	offset int
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		tasks: []*Task{},
	}
}

func (tq *TaskQueue) push(task *Task) {
	tq.Lock()
	defer tq.Unlock()

	tq.tasks = append(tq.tasks, task)
}

func (tq *TaskQueue) next() *Task {
	tq.Lock()
	defer tq.Unlock()

	if tq.offset >= len(tq.tasks) {
		return nil
	}

	defer func() { tq.offset++ }()
	return tq.tasks[tq.offset]
}

func (tq *TaskQueue) list() []*Task {
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
// 1. 处理同一个Domain下的优先级关系
// 2. 管理cookie等
type DomainBrowser struct {
	domain string

	priority *TaskQueue
	normal   *TaskQueue
}

func NewDomainBrowser(host string) *DomainBrowser {
	return &DomainBrowser{
		domain:   host,
		priority: NewTaskQueue(),
		normal:   NewTaskQueue(),
	}
}

func (d *DomainBrowser) Push(priority bool, task *Task) {
	if priority {
		d.priority.push(task)
	} else {
		d.normal.push(task)
	}
}

func (d *DomainBrowser) Next() *Task {
	if task := d.priority.next(); task != nil {
		return task
	}
	return d.normal.next()
}

func (d *DomainBrowser) List() []*Task {
	res := make([]*Task, 0, len(d.priority.list())+len(d.normal.list()))
	res = append(res, d.priority.list()...)
	res = append(res, d.normal.list()...)
	return res
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
	// 退避
	sh := backoff.NewBackOff(
		backoff.WithMinDelay(2*time.Second),
		backoff.WithMaxDelay(15*time.Second),
		backoff.WithFactor(2),
	)

	for {
		select {
		case <-ctx.Done():
			logx.Infof("[schedule] The process-%d[%s] will stop", index, d.domain)
			wg.Done()
			return
		default:
			task := d.Next()
			if task == nil {
				// 退避
				sh.SleepCtx(ctx)
				continue
			}

			if task.ID == 0 {
				task.ID = rand.Uint64()
			}
			logx.Infof("[schedule] The schedule[%x] begin to run, url: %s", task.ID, task.url)
			task.state = TaskStateRunning
			if err := task.Run(ctx, httpClient); err != nil {
				task.SetState(TaskStateFail)
				logx.Warnf("[schedule] The schedule[%x] run failed(try again after 5s): %v", task.ID, err)
				// 重试
				d.Push(true, task)
				time.Sleep(time.Second * 5)
				continue
			}
			sh.Reset()
			task.SetState(TaskStateSuccess)
			logx.Infof("[schedule] The schedule[%x] done.", task.ID)
			// 成功时的延迟
			time.Sleep(time.Second * 5)
		}
	}
}

func (d *DomainBrowser) reFail(t *Task) {

}

func (d *DomainBrowser) close() {
	// todo: close all task queue of the domain
}

func (d *DomainBrowser) MarshalJSON() ([]byte, error) {
	return nil, nil
}
