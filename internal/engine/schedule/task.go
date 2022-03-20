package schedule

import (
	"context"
	"fmt"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/monker-king/internal/utils"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"
)

// 请求完成后回调
type callback func(req *http.Request, resp *http.Response) error

type Task struct {
	ID    uint64
	Name  string
	Meta  map[string]interface{}
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

func init() {
	cookieJar, _ := cookiejar.New(nil)
	httpClient.Jar = cookieJar
}

func NewTask(name string, u *url.URL, meta map[string]interface{}, fun callback) *Task {
	return &Task{
		ID:    0,
		Name:  name,
		Meta:  meta,
		url:   u,
		state: TaskStateKnown,
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
	req.Close = true

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
