package view

import "github.com/gdamore/tcell/v2"

type ActionHandler func(key *tcell.EventKey) *tcell.EventKey

type KeyAction struct {
	Action ActionHandler
}

func (ui *AppUI) bindKeys() {

}
