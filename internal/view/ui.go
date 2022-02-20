package view

import (
	"context"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/yougtao/goutils/logx"
	"github.com/yougtao/monker-king/internal/engine/task"
)

type AppUI struct {
	app *tview.Application

	input *InputWrap
	page  *TaskPage

	actions map[*tcell.EventKey]*KeyAction
}

func NewUI() *AppUI {
	app := tview.NewApplication()
	return &AppUI{
		app:     app,
		actions: map[*tcell.EventKey]*KeyAction{},
	}
}

func (ui *AppUI) Init(fun func(string) error) {
	indicator := tview.NewTextView().SetText("Hello Monkey King")
	indicator.SetTitleAlign(tview.AlignCenter)

	ui.input = NewInputWrap(fun)
	ui.page = NewTaskPage()

	main := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(indicator, 1, 1, false).
		AddItem(ui.input, 1, 1, true).
		AddItem(ui.page, 1, 0, false)

	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if action, ok := ui.getAction(event); ok {
			return action.Action(event)
		}
		next := ui.input.InputHandler()
		next(event, nil)
		return nil
	})

	ui.app.SetRoot(main, true).SetFocus(ui.input)
}

func (ui *AppUI) Run(ctx context.Context) {
	if err := ui.app.Run(); err != nil {
		logx.Fatalf("[engine] application panic")
	}
}

func (ui *AppUI) getAction(key *tcell.EventKey) (*KeyAction, bool) {
	action, ok := ui.actions[key]
	return action, ok
}

func (ui *AppUI) AddTaskRow(t *task.Task) {
	ui.page.AddRow(t)
}
