package collector

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/monker-king/internal/engine/schedule/task"
	"github.com/xiaorui77/monker-king/internal/engine/types"
	"golang.org/x/net/html"
)

type HTMLElement struct {
	task      *task.Task
	Collector *Collector
	Request   *types.RequestWrap
	Response  *types.ResponseWarp

	Doc   *goquery.Document
	DOM   *goquery.Selection
	Index int
	Node  *html.Node
	Attr  []html.Attribute
}

// NewHTMLElement 创建可操作的HTML结构
func NewHTMLElement(t *task.Task, collector *Collector, resp *types.ResponseWarp, doc *goquery.Document, DOM *goquery.Selection, node *html.Node, index int) *HTMLElement {
	return &HTMLElement{
		task:      t,
		Collector: collector,
		Request:   resp.Request,
		Response:  resp,

		Doc:   doc,
		DOM:   DOM,
		Index: index,
		Node:  node,
		Attr:  node.Attr,
	}
}

func (e *HTMLElement) Visit(name, u string, resetDepth bool) error {
	logx.Infof("[Parsing] Task[%x] continue Visit url: %v", e.task.ID, u)
	URL, err := e.Request.URL.Parse(u)
	if err != nil {
		return err
	}
	// 片段信息置空, 片段信息即url中#后的内容
	URL.Fragment = ""
	if URL.Scheme == "//" {
		URL.Scheme = e.Request.URL.Scheme
	}
	logx.Infof("[parsing] Task[%x] add sub task: %v", e.task.ID, URL.String())
	return e.Collector.visit(e.task, name, URL.String(), resetDepth)
}

func (e *HTMLElement) GetText(selector, def string) string {
	if str := e.Doc.Find(selector).Text(); str != "" {
		return html.UnescapeString(str)
	}
	return def
}
