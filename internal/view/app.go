package view

import (
	"context"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/yougtao/goutils/logx"
	"github.com/yougtao/monker-king/internal/engine/schedule"
	"github.com/yougtao/monker-king/internal/engine/types"
	"github.com/yougtao/monker-king/internal/view/model"
	"os"
)

type AppUI struct {
	// app core layer
	collector types.Collect

	app  *tview.Application
	main *tview.Flex

	indicator *tview.TextView
	input     *InputWrap
	content   *PageStack

	actions map[tcell.Key]*KeyAction
}

func NewUI(collector types.Collect) *AppUI {
	return &AppUI{
		collector: collector,
		app:       tview.NewApplication(),
		actions:   map[tcell.Key]*KeyAction{},
	}
}

func (ui *AppUI) Init(taskData model.DataProducer) {
	ui.main = tview.NewFlex().SetDirection(tview.FlexRow)

	ui.indicator = tview.NewTextView()
	ui.indicator.SetTextAlign(tview.AlignCenter)
	ui.indicator.SetText("Hello Monkey King")

	ui.input = NewInputWrap(ui, ui.collector.Visit)
	ui.input.Init()
	ui.content = NewPageStack(ui)
	ui.content.Init(taskData)

	ui.main.AddItem(ui.indicator, 1, 1, false)
	ui.main.AddItem(ui.input, 0, 0, false)
	ui.main.AddItem(ui.content, 0, 10, false)

	ui.bindKeys()
	ui.app.SetInputCapture(ui.keyboardHandler)

	ui.app.SetRoot(ui.main, true)
}

func (ui *AppUI) Run(_ context.Context) {
	if err := ui.app.Run(); err != nil {
		logx.Fatalf("[engine] application panic")
	}
}

// BailOut exists the application.
func (ui *AppUI) BailOut() {
	// todo: stop main lookup
	ui.app.Stop()
	os.Exit(0)
}

func (ui *AppUI) GotoPage(name string) {
	ui.content.ChangePage(name)
}

// children pages

func (ui *AppUI) ResetPrompt() {
	ui.app.SetFocus(ui.input)
}

func (ui *AppUI) AddTaskRow(t *schedule.Task) {
	// ui.content.AddRow(t)
}

// AsKey converts rune to keyboard key.
func AsKey(evt *tcell.EventKey) tcell.Key {
	if evt.Key() != tcell.KeyRune {
		return evt.Key()
	}
	key := tcell.Key(evt.Rune())
	if evt.Modifiers() == tcell.ModAlt {
		key = tcell.Key(int16(evt.Rune()) * int16(evt.Modifiers()))
	}
	return key
}
