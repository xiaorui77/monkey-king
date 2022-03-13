package manager

import (
	"fmt"
	"github.com/yougtao/goutils/httpr"
)

func (m *Manager) HandleAddTask(c *httpr.Context) {
	data := &TaskRequest{}
	if err := c.ParseJSON(data); err != nil {
		c.ResultError(err)
		return
	}

	c.ResultMessage(fmt.Sprintf("add task success"), m.collector.Visit(data.Url))
}

func (m *Manager) HandleListTask(c *httpr.Context) {
	c.ResultData(m.collector.GetDataProducer().GetRows(), nil)
}
