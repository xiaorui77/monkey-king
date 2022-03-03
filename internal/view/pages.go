package view

import (
	"github.com/rivo/tview"
	"github.com/yougtao/monker-king/internal/view/model"
)

const TaskPageName = "Tasks"
const LogsPageName = "Logs"

type PageStack struct {
	*tview.Pages
	app *AppUI

	*Stack

	pages []Component
}

func NewPageStack(app *AppUI) *PageStack {
	return &PageStack{
		app: app,
	}
}

func (p *PageStack) Init(taskData model.DataProducer) {
	p.Pages = tview.NewPages()
	p.Pages.SetBorder(true)

	task := NewTaskPage(p.app, taskData)
	task.Init()
	p.AddPage(TaskPageName, task, true, true)
	p.pages = append(p.pages, task)

	logs := NewLogsPage(p.app)
	logs.Init()
	p.AddPage(LogsPageName, logs, true, true)
	p.pages = append(p.pages, logs)

	p.ChangePage(TaskPageName)
}

func (p *PageStack) ChangePage(name string) {
	page := p.GetPage(name)
	if page != nil {
		p.SwitchToPage(page.Name())
		p.SetTitle(" " + page.Name() + " ")

		p.notify(page)
	}
}

func (p *PageStack) GetPage(name string) Component {
	for _, page := range p.pages {
		if page.Name() == name {
			return page
		}
	}
	return nil
}

func (p *PageStack) notify(c Component) {
	c.Start()
}
