package engine

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/monker-king/internal/engine/task"
	"golang.org/x/net/html"
	"net/url"
	"strings"
)

// Response is the representation of a HTTP response made by a Collector
type Response struct {
	// StatusCode is the status code of the Response
	StatusCode int
	// Body is the content of the Response
	Body    []byte
	Request *Request
}

type Request struct {
	collector *Collector
	baseURL   *url.URL
	URL       *url.URL
}

// AbsoluteURL 根据相对路径获取完整url
func (r *Request) AbsoluteURL(u string) string {
	if strings.HasPrefix(u, "#") {
		return ""
	}
	var base *url.URL
	if r.baseURL != nil {
		base = r.baseURL
	} else {
		base = r.URL
	}
	absURL, err := base.Parse(u)
	if err != nil {
		return ""
	}
	absURL.Fragment = ""
	if absURL.Scheme == "//" {
		absURL.Scheme = r.URL.Scheme
	}
	return absURL.String()
}

type HTMLElement struct {
	task     *task.Task
	Request  *Request
	Response *Response

	Doc   *goquery.Document
	DOM   *goquery.Selection
	Index int
	Node  *html.Node
	Attr  []html.Attribute
}

// NewHTMLElement 创建可操作的HTML结构
func NewHTMLElement(t *task.Task, resp *Response, doc *goquery.Document, DOM *goquery.Selection, node *html.Node, index int) *HTMLElement {
	return &HTMLElement{
		task:     t,
		Request:  resp.Request,
		Response: resp,

		Doc:   doc,
		DOM:   DOM,
		Index: index,
		Node:  node,
		Attr:  node.Attr,
	}
}

func (e *HTMLElement) Visit(u string) error {
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
	return e.Request.collector.visit(e.task, URL)
}

func (e *HTMLElement) GetText(selector, def string) string {
	if str := e.Doc.Find(selector).Text(); str != "" {
		return html.UnescapeString(str)
	}
	return def
}
