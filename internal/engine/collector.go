package engine

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/yougtao/monker-king/internal/engine/task"
	"github.com/yougtao/monker-king/internal/utils"
	"io/ioutil"
	"net/http"
	"sync"
)

type Collector struct {
	store Store
	tasks task.Runner

	// 抓取成功后回调
	register sync.Mutex
	// 获取到页面后的回调, 为了保证顺序, 所以采用list
	htmlCallbacks []HtmlCallbackContainer
}

func NewCollector() *Collector {
	return &Collector{
		store:         NewStore(),
		tasks:         task.NewRunner(),
		htmlCallbacks: nil,
	}
}

func (c *Collector) Run(ctx context.Context) {
	c.tasks.Run(ctx)
}

func (c *Collector) Visit(url string) error {
	return c.scrape(context.TODO(), url, http.MethodGet, 1)
}

func (c *Collector) OnHTML(selector string, fun HtmlCallback) *Collector {
	c.register.Lock()
	defer c.register.Unlock()
	if c.htmlCallbacks == nil {
		c.htmlCallbacks = []HtmlCallbackContainer{}
	}
	c.htmlCallbacks = append(c.htmlCallbacks, HtmlCallbackContainer{selector, fun})
	return c
}

// 抓取网页, 目前仅支持GET
func (c *Collector) scrape(ctx context.Context, url, method string, depth int) error {
	if c.store.IsVisited(url) {
		return nil
	}

	hdr := http.Header{utils.UserAgentKey: []string{utils.RandomUserAgent()}}

	// 回调
	callback := func(req *http.Request, resp *http.Response) {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.Debugf("scrape html failed: %v", err)
			return
		}

		response := &Response{
			StatusCode: resp.StatusCode,
			Body:       body,
			Request: &Request{
				collector: c,
				baseURL:   req.URL, // todo: 该怎么设置
				URL:       req.URL,
			},
			Ctx: ctx,
		}

		// 通过task下载get到页面后通过回调执行
		logrus.Debugf("[scrape] 下载完成, handle callback handleOnHtml(%v)", response.Request.URL)
		c.handleOnHtml(response)
	}

	logrus.Debugf("[scrape] add Parser Task: %v", url)
	c.tasks.AddTask(task.NewParserTask(ctx, url, hdr, callback))
	return nil
}
