package view

import "github.com/rivo/tview"

type LogsPage struct {
	*tview.Flex
	app *AppUI

	indicator *tview.TextView
	logs      *tview.TextView
}

func NewLogsPage(app *AppUI) *LogsPage {
	return &LogsPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app,
	}
}

func (l *LogsPage) Init() {

}

func (l *LogsPage) Name() string {
	return "Logs"
}
