package manager

import (
	"fmt"
	"github.com/xiaorui77/goutils/httpr"
)

func (m *Manager) HandleAddTask(c *httpr.Context) {
	data := &TaskRequest{}
	if err := c.ParseJSON(data); err != nil {
		c.ResultError(err)
		return
	}

	c.ResultMessage(fmt.Sprintf("add task success: %v", data.Url), m.collector.Visit(data.Url))
}

func (m *Manager) HandleDeleteTask(c *httpr.Context) {
	data := &TaskRequest{}
	if err := c.ParseJSON(data); err != nil {
		c.ResultError(err)
		return
	}

	if m.collector.TaskManager().DeleteTask("", data.Id) {
		c.ResultMessage(fmt.Sprintf("delete task success: %v", data.Url), nil)
	} else {
		c.ResultError(fmt.Errorf("not found"))
	}
}

func (m *Manager) HandleBrowserTree(c *httpr.Context) {
	domain, ok := c.Params["domain"]
	if !ok {
		return
	}
	c.ResultData(m.collector.TaskManager().GetTree(domain), nil)
}

func (m *Manager) HandleListTask(c *httpr.Context) {
	c.ResultData(m.collector.GetDataProducer().GetRows(), nil)
}
