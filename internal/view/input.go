package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

type InputWrap struct {
	*tview.InputField
}

func NewInputWrap(callback func(string) error) *InputWrap {
	input := tview.NewInputField()
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			str := strings.TrimSpace(input.GetText())
			if str != "" && callback(str) == nil {
				input.SetText("")
			}
		}
	})

	return &InputWrap{input}
}
