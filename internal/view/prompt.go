package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

const (
	ModeNode = iota
	ModeCmd
	ModeInput
)

type InputWrap struct {
	*tview.InputField
	app *AppUI

	active   bool
	mode     int
	callback func(string) error
}

func NewInputWrap(app *AppUI, callback func(string) error) *InputWrap {
	return &InputWrap{
		InputField: tview.NewInputField(),
		app:        app,
		callback:   callback,
	}
}

func (i *InputWrap) Init() {
	i.SetBorder(true)
	i.SetBackgroundColor(tcell.ColorDefault)

	i.SetInputCapture(i.keyboard)

	i.SetDoneFunc(i.OnComplete)
}

func (i *InputWrap) Active(activate bool, mode int) {
	if activate {
		i.active = true
		if i.mode != mode {
			i.mode = mode
			i.SetText("")
		}
		i.app.app.SetFocus(i)
		i.app.main.ResizeItem(i, 3, 1)
		return
	}

	i.active = false
	i.app.app.SetFocus(i.app.content)
	i.app.main.ResizeItem(i, 0, 0)
}

// IsActivated returns true if command is active, false otherwise.
func (i *InputWrap) IsActivated() bool {
	return i.active
}

func (i *InputWrap) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	switch evt.Key() {
	case tcell.KeyEsc:
		i.Active(false, ModeNode)
	case tcell.KeyEnter:
		return evt
	case tcell.KeyRune:
		return evt
	}
	return evt
}

func (i *InputWrap) OnComplete(key tcell.Key) {
	if key != tcell.KeyEnter || i.GetText() == "" {
		return
	}

	if i.mode == ModeCmd {
		i.OnCompleteCmd()
	} else if i.mode == ModeInput {
		i.OnCompleteInput()
	}
}

func (i *InputWrap) OnCompleteCmd() {
	str := strings.TrimSpace(i.GetText())
	i.app.content.ChangePage(str, true)
	i.SetText("")
	i.Active(false, ModeNode)
}

func (i *InputWrap) OnCompleteInput() {
	str := strings.TrimSpace(i.GetText())
	if str != "" && i.app.collector.Visit(nil, str) == nil {
		i.SetText("")
		i.Active(false, ModeNode)
	}
}
