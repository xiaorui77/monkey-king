package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/yougtao/monker-king/internal/engine/task"
	"strconv"
)

type TaskPage struct {
	*tview.Table
}

func NewTaskPage() *TaskPage {
	table := tview.NewTable()
	table.SetSelectable(true, false)

	table.SetCell(0, 0, &tview.TableCell{
		Text:            "Name",
		Color:           tcell.ColorGreen,
		BackgroundColor: tcell.ColorFireBrick,
	})
	table.SetCell(0, 1, &tview.TableCell{
		Text:            "URL",
		Color:           tcell.ColorGreen,
		BackgroundColor: tcell.ColorFireBrick,
	})

	table.SetFixed(1, 0)
	return &TaskPage{
		table,
	}
}

func (t *TaskPage) AddRow(task *task.Task) {
	r := t.GetRowCount()
	t.SetCell(r, 0, &tview.TableCell{
		Text: strconv.FormatUint(task.ID, 10),
	})
	t.SetCell(r, 1, &tview.TableCell{
		Text: task.Url.Path,
	})
}
