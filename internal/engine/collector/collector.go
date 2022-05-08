package collector

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/monker-king/internal/config"
	"github.com/xiaorui77/monker-king/internal/engine/schedule"
	"github.com/xiaorui77/monker-king/internal/engine/schedule/api"
	"github.com/xiaorui77/monker-king/internal/engine/schedule/task"
	"github.com/xiaorui77/monker-king/internal/engine/types"
	"github.com/xiaorui77/monker-king/internal/storage"
	"github.com/xiaorui77/monker-king/internal/utils/fileutil"
	"github.com/xiaorui77/monker-king/internal/view/model"
	"net/url"
	"sync"
)

type Collector struct {
	config    *config.Config
	scheduler *schedule.Scheduler
	store     storage.Store
	storage   storage.Storage

	// visited list
	visitedList  map[string]bool
	visitedMutex sync.Mutex

	// 抓取成功后回调
	register sync.Mutex

	// HTML 回调
	htmlCallbacks    []HtmlCallbackContainer
	ResponseCallback []ResponseCallback
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

	c := &Collector{
		config:  config,
		store:   store,
		storage: storage.NewStorage("192.168.17.1:3306"),

		visitedList:   map[string]bool{},
		htmlCallbacks: nil,
	}
	c.scheduler = schedule.NewRunner(c, c.storage)
	return c, nil
}

func (c *Collector) Run(ctx context.Context) {
	logx.Infof("[collector] The Collector already running...")
	c.scheduler.Run(ctx)
	logx.Infof("[collector] The Collector has been stopped")
}

func (c *Collector) TaskManager() api.TaskManage {
	return c.scheduler
}

// Visit 是对外的接口, 可以访问指定url
func (c *Collector) Visit(rawUrl string) error {
	logx.Infof("[collector] Visit url: %v", rawUrl)
	if len(rawUrl) == 0 {
		return errors.New("rawUrl is empty")
	}

	if _, err := url.Parse(rawUrl); err != nil {
		logx.Warnf("[collector] new schedule failed with parse url(%v): %v", rawUrl, err)
		return err
	}
	return c.visit(nil, "", rawUrl, true)
}

// Download 下载保存, todo: 移动到parsing中
func (c *Collector) Download(t *task.Task, name, path string, urlRaw string) error {
	if _, err := url.Parse(urlRaw); err != nil {
		logx.Warnf("[schedule] new schedule failed with parse url(%v): %v", urlRaw, err)
		return errors.New("未能识别的URL")
	}
	return c.scheduler.AddTask(task.NewTask(name, t, urlRaw, c.save).
		SetPriority(1).SetMeta(task.MetaSavePath, path).SetMeta("save_name", name))
}

func (c *Collector) visit(parent *task.Task, name, url string, resetDepth bool) error {
	if err := c.filter(url); err != nil {
		logx.Warnf("[collector] filter url(%s) cause by: %v", url, err)
		return err
	}
	t := task.NewTask(name, parent, url, c.parsing, task.AddOnCreatedHandler(
		func(task *task.Task) {
			if resetDepth {
				task.Depth = 0
			}
		}))

	return c.AddTask(t)
}

func (c *Collector) AddTask(t *task.Task) error {
	if t == nil {
		return nil
	}
	logx.Debugf("[scrape] add parsed Task:%v", t.String())
	return c.scheduler.AddTask(t)
}

// 回调函数: 处理抓取到的页面
func (c *Collector) parsing(task *task.Task, resp *types.ResponseWarp) error {
	logx.Debugf("[collector] Task[%08x] parsing response", task.ID)
	c.handleOnHtml(task, resp)
	c.recordVisit(resp.Request.URL.String())
	logx.Infof("[collector] Task[%08x] parsing and handle done.", task.ID)
	return nil
}

// 回调函数: 保存文件
func (c *Collector) save(t *task.Task, resp *types.ResponseWarp) error {
	name := t.Meta[task.MetaSaveName].(string)
	path := t.Meta[task.MetaSavePath].(string)

	logx.Infof("[collector] Task[%x] save file \"%s\" to: %s", t.ID, name, path)
	return fileutil.SaveImage(resp.Body, path, name)
}

// 借些页面, 处理回调
func (c *Collector) handleOnHtml(task *task.Task, resp *types.ResponseWarp) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body))
	if err != nil {
		logx.Debugf("parse html to document failed: %v", err)
		return
	}
	for _, callback := range c.htmlCallbacks {
		index := 1
		doc.Find(callback.Selector).Each(func(_ int, selection *goquery.Selection) {
			for _, node := range selection.Nodes {
				e := NewHTMLElement(task, c, resp, doc, selection, node, index)
				index++
				callback.fun(task, e)
			}
		})
	}
}

// @return ok: 是否继续
func (c *Collector) filter(url string) error {
	if c.isVisited(url) {
		return fmt.Errorf("the URL has been browsed")
	}
	return nil
}

func (c *Collector) recordVisit(url string) {
	c.visitedMutex.Lock()
	defer c.visitedMutex.Unlock()
	if c.config.Persistent {
		c.store.Visit(url)
	}
	c.visitedList[url] = true
}

func (c *Collector) isVisited(url string) bool {
	c.visitedMutex.Lock()
	defer c.visitedMutex.Unlock()
	b, ok := c.visitedList[url]
	if ok {
		return b
	}
	return false
}

func (c *Collector) GetDataProducer() model.DataProducer {
	return c.scheduler
}
