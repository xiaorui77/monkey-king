package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
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
		Pages: tview.NewPages(),
		app:   app,
		Stack: NewStack(),
	}
}

func (p *PageStack) Init() {
	p.Pages.SetBackgroundColor(tcell.ColorDefault)

	task := NewTaskPage(p.app, p.app.collector.GetDataProducer())
	task.Init()
	p.pages = append(p.pages, task)

	logs := NewLogsPage(p.app)
	logs.Init()
	p.pages = append(p.pages, logs)

	p.Stack.AddListener(p)

	// default page
	p.ChangePage(TaskPageName, true)
}

func (p *PageStack) ChangePage(name string, clearStack bool) {
	page := p.GetPage(name)
	if page != nil {
		if clearStack {
			p.Clear()
		}
		p.Push(page)
	}
}

func (p *PageStack) GetPage(name string) Component {
	for _, page := range p.pages {
		if strings.ToLower(page.Name()) == strings.ToLower(name) {
			return page
		}
	}
	return nil
}

// StackPushed notifies a new component was pushed.
func (p *PageStack) StackPushed(c Component) {
	p.Pages.AddPage(c.Name(), c, true, true)
	p.Pages.ShowPage(c.Name())

	c.Start()
}

// StackPopped notifies a component was removed.
func (p *PageStack) StackPopped(c, o Component) {
	p.Pages.RemovePage(c.Name())
}
