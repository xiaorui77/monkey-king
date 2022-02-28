package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

type InputWrap struct {
	*tview.InputField
	app *AppUI

	model    bool
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

	i.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			str := strings.TrimSpace(i.GetText())
			if str != "" && i.callback(str) == nil {
				i.SetText("")
			}
		}
	})
}

func (i *InputWrap) Active(activate bool) {
	if activate {
		i.model = true
		i.app.app.SetFocus(i)
		i.app.main.ResizeItem(i, 3, 1)
		return
	}

	i.model = false
	i.app.app.SetFocus(i.app.content)
	i.app.main.ResizeItem(i, 0, 0)
}

func (i *InputWrap) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	switch evt.Key() {
	case tcell.KeyEsc:
		i.Active(false)
	case tcell.KeyRune:
		return evt
	}
	return nil
}

// InCmdMode returns true if command is active, false otherwise.
func (i *InputWrap) InCmdMode() bool {
	return i.model
}
