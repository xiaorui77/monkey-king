package engine

import "github.com/xiaorui77/monker-king/internal/engine/task"

type HtmlCallback func(task *task.Task, element *HTMLElement)
type HtmlCallbackContainer struct {
	Selector string
	fun      HtmlCallback
}

// OnHTMLAny 会对匹配到的每一个元素分别执行回调操作
func (c *Collector) OnHTMLAny(selector string, fun HtmlCallback) *Collector {
	c.register.Lock()
	defer c.register.Unlock()
	if c.htmlCallbacks == nil {
		c.htmlCallbacks = []HtmlCallbackContainer{}
	}
	c.htmlCallbacks = append(c.htmlCallbacks, HtmlCallbackContainer{selector, fun})
	return c
}
