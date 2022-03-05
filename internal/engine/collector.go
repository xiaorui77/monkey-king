package engine

import (
	"context"
	"errors"
	"fmt"
	"github.com/yougtao/goutils/logx"
	"github.com/yougtao/monker-king/internal/config"
	"github.com/yougtao/monker-king/internal/engine/schedule"
	"github.com/yougtao/monker-king/internal/storage"
	"github.com/yougtao/monker-king/internal/view/model"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
)

type Collector struct {
	config    *config.Config
	store     storage.Store
	scheduler schedule.Runner

	// visited list
	visitedList map[string]bool

	// 抓取成功后回调
	register sync.Mutex
	// 获取到页面后的回调, 为了保证顺序, 所以采用list
	htmlCallbacks []HtmlCallbackContainer
}

func NewCollector(config *config.Config) (*Collector, error) {
	store, err := storage.NewRedisStore("127.0.0.1:6379")
	if err != nil {
		logx.Errorf("new collector failed: %v", err)
		return nil, errors.New("connect redis failed")
	}

	runner := schedule.NewRunner(store)
	c := &Collector{
		config:    config,
		store:     store,
		scheduler: runner,

		visitedList:   map[string]bool{},
		htmlCallbacks: nil,
	}
	return c, nil
}

func (c *Collector) Run(ctx context.Context) {
	c.scheduler.Run(ctx)
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
func (c *Collector) scrape(ctx context.Context, urlRaw, method string, depth int) error {
	if c.isVisited(urlRaw) {
		return nil
	}

	// 回调
	callback := func(req *http.Request, resp *http.Response) error {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logx.Debugf("scrape html failed: %v", err)
			return fmt.Errorf("scrape html failed")
		}

		response := &Response{
			StatusCode: resp.StatusCode,
			Body:       body,
			Request: &Request{
				collector: c,
				baseURL:   req.URL, // todo: 该怎么设置
				URL:       req.URL,
			},
		}

		// 通过task下载get到页面后通过回调执行
		logx.Debugf("[scrape] 下载完成, handle callback handleOnHtml[%v]", urlRaw)
		c.handleOnHtml(response)
		c.recordVisit(urlRaw)
		logx.Debugf("[scrape] 分析完成, handleOnHtml[%v]", urlRaw)
		return nil
	}

	u, err := url.Parse(urlRaw)
	if err != nil {
		logx.Warnf("[schedule] new schedule failed with parse url(%v): %v", urlRaw, err)
		return errors.New("未能识别的URL")
	}
	c.AddTask(schedule.NewTask(u, callback))
	return nil
}

func (c *Collector) AddTask(t *schedule.Task) {
	if t == nil {
		return
	}
	logx.Debugf("[scrape] add Parser Task: %v", t.Url.Path)
	// c.ui.AddTaskRow(t)
	c.scheduler.AddTask(t, false)
}

// 对抓取到的页面回调
func (c *Collector) onScrape(req *http.Request, resp *http.Response) error {
	// todo: 代迁移
	return nil
}

func (c *Collector) recordVisit(url string) {
	if c.config.Persistent {
		c.store.Visit(url)
	}
	c.visitedList[url] = true
}

func (c *Collector) isVisited(url string) bool {
	b, ok := c.visitedList[url]
	if ok {
		return b
	}
	return false
}

func (c *Collector) GetDataProducer() model.DataProducer {
	d, _ := c.scheduler.(*schedule.Scheduler)
	return d
}
