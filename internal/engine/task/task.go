package task

import (
	"fmt"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/monker-king/internal/utils/domain"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// 请求完成后回调
type callback func(task *Task, req *http.Request, resp *http.Response) error

type OnResponse func(req *http.Request, resp *http.Response)

type OnResponseError func(resp *http.Response, err error)

const (
	StateUnknown = iota
	StateRunning
	StateInit
	StateFail
	StateSuccess
)

var StateStatus = map[int]string{
	0: "known",
	1: "running",
	2: "init",
	3: "Fail",
	4: "Success",
}

type Task struct {
	ID       uint64   `json:"id"`
	ParentId uint64   `json:"parentId"`
	Parent   *Task    `json:"-"`
	Depth    int      `json:"depth"`
	Name     string   `json:"name"`
	State    int      `json:"state"`
	Url      *url.URL `json:"url"`
	Domain   string   `json:"domain"`

	// 优先级: [0, MAX_INT), 值越大优先级越高
	Priority int `json:"priority"`

	Meta map[string]interface{} `json:"meta"`
	Time time.Time              `json:"time"`

	// 主回调函数, 后续考虑优化合并
	callback                callback
	onResponseHandlers      []OnResponse
	onResponseErrorHandlers []OnResponseError
}

func NewTask(name string, parent *Task, u *url.URL, meta map[string]interface{}, fun callback) *Task {
	t := &Task{
		ID:       rand.Uint64(),
		Name:     name,
		Meta:     meta,
		Url:      u,
		State:    StateUnknown,
		Time:     time.Now(),
		callback: fun,
	}
	if parent != nil {
		t.Domain = parent.Domain
		t.ParentId = parent.ID
		t.Parent = parent
		t.Depth = parent.Depth + 1
	} else {
		t.Domain = domain.CalDomain(u)
	}
	return t
}

func (t *Task) SetPriority(p int) *Task {
	t.Priority = p
	return t
}

func (t *Task) ResetDepth() *Task {
	t.Depth = 0
	return t
}

func (t *Task) HandleOnResponse(req *http.Request, resp *http.Response) {
	for _, handler := range t.onResponseHandlers {
		handler(req, resp)
	}

	if resp.StatusCode == http.StatusOK {
		if err := t.callback(t, req, resp); err != nil {
			logx.Debugf("[schedule] Task[%x] failed: %v", t.ID, err)
			t.SetState(StateFail)
		} else {
			logx.Debugf("[schedule] Task[%x] success.", t.ID)
			t.SetState(StateSuccess)
		}
	} else {
		logx.Debugf("[schedule] The task[%x] failed with unknown status code[%d]", t.ID, resp.StatusCode)
		t.SetState(StateFail)
	}
}

func (t *Task) HandleOnResponseErr(resp *http.Response, err error) {
	for _, handler := range t.onResponseErrorHandlers {
		handler(resp, err)
	}
}

func (t *Task) SetState(state int) {
	t.State = state
}

func (t *Task) String() string {
	return fmt.Sprintf("[%s] %s", t.Name, t.Url.String())
}
