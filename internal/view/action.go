package view

import "github.com/gdamore/tcell/v2"

type ActionHandler func(key *tcell.EventKey) *tcell.EventKey

type KeyAction struct {
	Key         tcell.Key
	Action      ActionHandler
	Description string
}

func NewAction(key tcell.Key, action ActionHandler) *KeyAction {
	return &KeyAction{
		Key:    key,
		Action: action,
	}
}

func (ui *AppUI) keyboardHandler(event *tcell.EventKey) *tcell.EventKey {
	if action, ok := ui.GetAction(event); ok {
		return action.Action(event)
	}
	return event
}

func (ui *AppUI) bindKeys() {
	ui.actions[KeyColon] = NewAction(KeyColon, ui.activateCmd)
	ui.actions[tcell.KeyCtrlC] = NewAction(tcell.KeyCtrlC, ui.exitCmd)
}

func (ui *AppUI) GetAction(key *tcell.EventKey) (*KeyAction, bool) {
	action, ok := ui.actions[AsKey(key)]
	return action, ok
}

// 激活命令窗口, with the Key ":"
func (ui *AppUI) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if ui.input.InCmdMode() {
		return evt
	}
	ui.input.Active(true)
	return nil
}

// 退出命令
func (ui *AppUI) exitCmd(evt *tcell.EventKey) *tcell.EventKey {
	ui.BailOut()
	return nil
}

// Defines char keystrokes.
const (
	KeyA tcell.Key = iota + 97
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
	KeyHelp  = 63
	KeySlash = 47
	KeyColon = 58
	KeySpace = 32
)
