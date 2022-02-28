package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/yougtao/monker-king/internal/engine/task"
	"strconv"
)

type TaskPage struct {
	*tview.Table

	data TableData
}

func NewTaskPage() *TaskPage {
	return &TaskPage{
		Table: tview.NewTable(),
	}
}

func (t *TaskPage) Init(data TableData) {
	t.Table.SetSelectable(true, false)
	t.SetCell(0, 0, &tview.TableCell{
		Text:            "Name",
		Color:           tcell.ColorGreen,
		BackgroundColor: tcell.ColorFireBrick,
	})
	t.SetCell(0, 1, &tview.TableCell{
		Text:            "URL",
		Color:           tcell.ColorGreen,
		BackgroundColor: tcell.ColorFireBrick,
	})
	t.SetFixed(1, 0)
}

func (t *TaskPage) Name() string {
	return "Task"
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
