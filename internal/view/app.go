package view

import (
	"context"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/monker-king/internal/engine/interfaces"
	"io"
	"os"
)

type AppUI struct {
	// app core layer
	collector interfaces.Collect

	app  *tview.Application
	main *tview.Flex

	indicator *tview.TextView
	input     *InputWrap
	content   *PageStack

	actions map[tcell.Key]*KeyAction
}

func NewUI(collector interfaces.Collect) *AppUI {
	return &AppUI{
		collector: collector,
		app:       tview.NewApplication(),
		actions:   map[tcell.Key]*KeyAction{},
	}
}

func (ui *AppUI) Init() {
	ui.main = tview.NewFlex().SetDirection(tview.FlexRow)
	ui.main.SetBackgroundColor(tcell.ColorDefault)

	ui.indicator = tview.NewTextView()
	ui.indicator.SetTextAlign(tview.AlignCenter)
	ui.indicator.SetText("Hello Monkey King")
	ui.indicator.SetBackgroundColor(tcell.ColorDefault)

	ui.input = NewInputWrap(ui, ui.collector)
	ui.input.Init()
	ui.content = NewPageStack(ui)
	ui.content.Init()

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

func (ui *AppUI) IsRunning() bool {
	// todo
	return true
}

// BailOut exists the application.
func (ui *AppUI) BailOut() {
	// todo: stop main lookup
	ui.app.Stop()
	os.Exit(0)
}

func (ui *AppUI) GotoPage(name string) {
	ui.content.ChangePage(name, true)
}

// children pages

func (ui *AppUI) ResetPrompt() {
	ui.app.SetFocus(ui.input)
}

func (ui *AppUI) GetLogsWriter() io.Writer {
	return ui.content.GetPage(LogsPageName).(*LogsPage).GetModel()
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
