package task

import (
	"context"
	"fmt"
	"github.com/yougtao/goutils/logx"
	"github.com/yougtao/monker-king/internal/utils"
	"net/http"
	"net/url"
)

type Task struct {
	ID  uint64
	url *url.URL
	fun callback
}

func NewTask(urlRaw string, fun callback) *Task {
	u, err := url.Parse(urlRaw)
	if err != nil {
		logx.Warnf("[task] new task failed with parse url(%v): %v", urlRaw, err)
		return nil
	}
	return &Task{
		ID:  0,
		url: u,
		fun: fun,
	}
}

func (task *Task) Run(ctx context.Context, client *http.Client) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, task.url.String(), nil)
	if err != nil {
		logx.Warnf("[task] The task[%x] failed during the new request: %v", task.ID, err)
		return fmt.Errorf("new request fail: %v", err)
	}
	req.Header.Set(utils.UserAgentKey, utils.RandomUserAgent())

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		logx.Warnf("[task] The task[%x] failed during the do request: %v", task.ID, err)
		return fmt.Errorf("do request fail")
	}

	if resp.StatusCode == http.StatusOK {
		return task.fun(req, resp)
	} else {
		logx.Warnf("[task] The task[%x] failed with unknown status code[%d]", task.ID, resp.StatusCode)
		return fmt.Errorf("do request fail with status code[%d]", resp.StatusCode)
	}
}

// 请求完成后回调
type callback func(req *http.Request, resp *http.Response) error
