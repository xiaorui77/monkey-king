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
	i.SetInputCapture(i.keyboard)

	i.SetDoneFunc(i.OnComplete)
}

func (i *InputWrap) Active(activate bool, mode int) {
	if activate {
		i.active = true
		i.mode = mode
		i.app.app.SetFocus(i)
		i.app.main.ResizeItem(i, 3, 1)
		return
	}

	i.active = false
	i.mode = ModeNode
	i.app.app.SetFocus(i.app.content)
	i.app.main.ResizeItem(i, 0, 0)
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
	if key == tcell.KeyEnter {
		str := strings.TrimSpace(i.GetText())
		if str != "" && i.callback(str) == nil {
			i.SetText("")
		}
	}
}

// IsActivated returns true if command is active, false otherwise.
func (i *InputWrap) IsActivated() bool {
	return i.active
}
