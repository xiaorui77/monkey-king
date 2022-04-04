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
	StateKnown = iota
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
	ID       uint64
	ParentId uint64
	Name     string
	State    int
	Domain   string
	Priority int
	Url      *url.URL

	Meta map[string]interface{}
	Time time.Time

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
		State:    StateKnown,
		Time:     time.Now(),
		callback: fun,
	}
	if parent != nil {
		t.Domain = parent.Domain
		t.ParentId = parent.ID
	} else {
		t.Domain = domain.CalDomain(u)
	}
	return t
}

func (t *Task) SetPriority(p int) *Task {
	t.Priority = p
	return t
}

func (t *Task) HandleOnResponse(req *http.Request, resp *http.Response) {
	for _, handler := range t.onResponseHandlers {
		handler(req, resp)
	}

	if resp.StatusCode == http.StatusOK {
		if err := t.callback(t, req, resp); err != nil {
			logx.Warnf("[schedule] Task[%x] failed: %v", t.ID, err)
			t.SetState(StateFail)
		} else {
			logx.Infof("[schedule] Task[%x] done.", t.ID)
			t.SetState(StateSuccess)
		}
	} else {
		logx.Warnf("[schedule] The task[%x] failed with unknown status code[%d]", t.ID, resp.StatusCode)
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
