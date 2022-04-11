package collector

import (
	"fmt"
	"github.com/xiaorui77/monker-king/internal/engine/task"
	"github.com/xiaorui77/monker-king/internal/engine/types"
	"github.com/xiaorui77/monker-king/pkg/error"
	"net/http"
)

// ResponseCallback is the callback function for response
type ResponseCallback func(resp *types.ResponseWarp)

func (c *Collector) HandleOnResponse(resp *types.ResponseWarp) error.Error {
	for _, handler := range c.ResponseCallback {
		handler(resp)
	}

	if resp.StatusCode != http.StatusOK {
		return &error.Err{Code: task.ErrHttpUnknown + resp.StatusCode, Err: fmt.Errorf("response code is not ok[%v]", resp.StatusCode)}
	}
	return nil
}

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
