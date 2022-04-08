package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/monker-king/internal/config"
	"github.com/xiaorui77/monker-king/internal/engine/schedule"
	"github.com/xiaorui77/monker-king/internal/engine/task"
	"github.com/xiaorui77/monker-king/internal/storage"
	"github.com/xiaorui77/monker-king/internal/utils/fileutil"
	"github.com/xiaorui77/monker-king/internal/view/model"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
)

type Collector struct {
	config    *config.Config
	scheduler *schedule.Scheduler
	store     storage.Store

	// visited list
	visitedList map[string]bool

	// 抓取成功后回调
	register sync.Mutex
	// 获取到页面后的回调, 为了保证顺序, 所以采用list
	htmlCallbacks []HtmlCallbackContainer
}

func NewCollector(config *config.Config) (*Collector, error) {
	var store storage.Store
	var err error
	if config.Persistent {
		store, err = storage.NewRedisStore("127.0.0.1:6379")
		if err != nil {
			logx.Errorf("new collector failed: %v", err)
			return nil, errors.New("connect redis failed")
		}
	}

	scheduler := schedule.NewRunner(store)
	c := &Collector{
		config:    config,
		store:     store,
		scheduler: scheduler,

		visitedList:   map[string]bool{},
		htmlCallbacks: nil,
	}
	return c, nil
}

func (c *Collector) Run(ctx context.Context) {
	logx.Infof("[collector] Already running...")
	c.scheduler.Run(ctx)
	logx.Infof("[collector] has been stopped")
}

func (c *Collector) Scheduler() *schedule.Scheduler {
	return c.scheduler
}

// Visit 是对外的接口, 可以访问指定url
func (c *Collector) Visit(parent *task.Task, rawUrl string) error {
	logx.Infof("[collector] Visit url: %v", rawUrl)
	if len(rawUrl) == 0 {
		return errors.New("rawUrl is empty")
	}

	u, err := url.Parse(rawUrl)
	if err != nil {
		logx.Warnf("[collector] new schedule failed with parse url(%v): %v", rawUrl, err)
		return err
	}
	return c.visit(parent, u)
}

// Download 下载保存, todo: 移动到parsing中
func (c *Collector) Download(t *task.Task, name, path string, urlRaw string) error {
	u, err := url.Parse(urlRaw)
	if err != nil {
		logx.Warnf("[schedule] new schedule failed with parse url(%v): %v", urlRaw, err)
		return errors.New("未能识别的URL")
	}
	c.scheduler.AddTask(task.NewTask(name, t, u, c.save).
		SetPriority(1).SetMeta("save_path", path).SetMeta("save_name", name))
	return nil
}

func (c *Collector) visit(parent *task.Task, u *url.URL) error {
	if len(u.Host) == 0 {
		logx.Warnf("[collector] visit url(%s) failed: rawUrl is invalid", u.String())
		return errors.New("rawUrl is invalid")
	}
	if err := c.filter(u); err != nil {
		logx.Warnf("[collector] filter url(%s) cause by: %v", u.String(), err)
		return err
	}

	c.AddTask(task.NewTask("", parent, u, c.scrape))
	return nil
}

func (c *Collector) AddTask(t *task.Task) {
	if t == nil {
		return
	}
	logx.Debugf("[scrape] add parsed Task:%v", t.String())
	c.scheduler.AddTask(t)
}

// 回调函数: 处理抓取到的页面
// todo: 对页面分类
func (c *Collector) scrape(task *task.Task, req *http.Request, resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logx.Debugf("[collector] scrape read body failed: %v", err)
		return fmt.Errorf("scrape read body failed")
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
	logx.Debugf("[collector] Task[%x] scrape done, begin handleOnHtml()", task.ID)
	c.handleOnHtml(task, response)
	c.recordVisit(req.URL.String())
	logx.Infof("[collector] Task[%x] scrape and handle done.", task.ID)
	return nil
}

// 回调函数: 保存文件
func (c *Collector) save(t *task.Task, _ *http.Request, resp *http.Response) error {
	defer func() {
		_ = resp.Body.Close()
	}()
	name := t.Meta["save_name"].(string)
	path := t.Meta["save_path"].(string)

	reader := &fileutil.VisualReader{
		Reader: resp.Body,
		Total:  resp.ContentLength,
	}
	bs, err := reader.ReadAll()
	if err != nil {
		t.SetMeta("reader", reader)
		return fmt.Errorf("reading resp.Body when[%v/%v] failed: %v", reader.Cur, reader.Total, err)
	}
	logx.Infof("[collector] Task[%x] save file \"%s\" to: %s", t.ID, name, path)
	return fileutil.SaveImage(bs, path, name)
}

// 借些页面, 处理回调
func (c *Collector) handleOnHtml(task *task.Task, resp *Response) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body))
	if err != nil {
		logx.Debugf("parse html to document failed: %v", err)
		return
	}
	for _, callback := range c.htmlCallbacks {
		index := 1
		doc.Find(callback.Selector).Each(func(_ int, selection *goquery.Selection) {
			for _, node := range selection.Nodes {
				e := NewHTMLElement(task, resp, doc, selection, node, index)
				index++
				callback.fun(task, e)
			}
		})
	}
}

// @return ok: 是否继续
func (c *Collector) filter(u *url.URL) error {
	if c.isVisited(u.String()) {
		return fmt.Errorf("the URL has been browsed")
	}
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
	return c.scheduler
}
