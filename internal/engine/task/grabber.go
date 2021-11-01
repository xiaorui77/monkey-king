package task

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

// 抓取页面类型任务
type grabber struct {
	url      string
	header   http.Header
	callback Callback
	ctx      context.Context
}

// NewParserTask 新建一个爬取任务
func NewParserTask(ctx context.Context, url string, hdr http.Header, f Callback) *grabber {
	return &grabber{
		url:      url,
		header:   hdr,
		callback: f,
		ctx:      ctx,
	}
}

func (p *grabber) Run(ctx context.Context, client *http.Client) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		logrus.Warnf("[task] new request failed: %v", err)
		return fmt.Errorf("new request failed")
	}
	req.Header = p.header

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		logrus.Infof("[task] do request failed: %v", err)
		return fmt.Errorf("do request failed")
	}

	// handle
	p.callback(req, resp)
	return nil
}

type Callback func(req *http.Request, resp *http.Response)
