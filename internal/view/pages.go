package view

import "github.com/rivo/tview"

const TaskPageName = "task"
const LogsPageName = "logs"

type PageStack struct {
	*tview.Pages
	app *AppUI

	*Stack

	pages []Primitive
}

func NewPageStack(app *AppUI) *PageStack {
	return &PageStack{
		app: app,
	}
}

func (p *PageStack) Init(taskData TableData) {
	p.Pages = tview.NewPages()
	p.Pages.SetBorder(true)

	task := NewTaskPage()
	task.Init(taskData)
	p.AddPage(TaskPageName, task, true, true)

	logs := NewLogsPage(p.app)
	logs.Init()
	p.AddPage(LogsPageName, logs, true, true)

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

func (p *PageStack) GetPage(name string) Primitive {
	for _, page := range p.pages {
		if page.Name() == name {
			return page
		}
	}
	return nil
}

func (p *PageStack) notify(c Primitive) {

}
