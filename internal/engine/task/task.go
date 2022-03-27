package task

import (
	"fmt"
	"github.com/xiaorui77/goutils/logx"
	"net/http"
	"net/url"
	"time"
)

// 请求完成后回调
type callback func(req *http.Request, resp *http.Response) error

type OnResponse func(req *http.Request, resp *http.Response)

type OnResponseError func(resp *http.Response, err error)

type Task struct {
	ID       uint64
	Name     string
	Meta     map[string]interface{}
	Priority int
	State    int
	Url      *url.URL
	Time     time.Time

	callback                callback
	onResponseHandlers      []OnResponse
	onResponseErrorHandlers []OnResponseError
}

const (
	StateKnown = iota
	StateRunning
	StateInit
	StateFail
	StateSuccess
)

var StateStatus = map[int]string{
	1: "running",
	2: "init",
	3: "Fail",
	4: "Success",
}

func NewTask(name string, u *url.URL, meta map[string]interface{}, fun callback) *Task {
	return &Task{
		ID:       0,
		Name:     name,
		Meta:     meta,
		Url:      u,
		State:    StateKnown,
		Time:     time.Now(),
		callback: fun,
	}
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
		if err := t.callback(req, resp); err != nil {
			t.SetState(StateFail)
		} else {
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
