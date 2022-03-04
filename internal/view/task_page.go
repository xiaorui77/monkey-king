package view

import (
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/yougtao/monker-king/internal/view/model"
	"strconv"
	"time"
)

const (
	HeaderID     = "ID"
	HeaderDomain = "Domain"
	HeaderName   = "Name"
	HeaderStatus = "State"
	HeaderAge    = "Age"
	HeaderURL    = "URL"
)

type TaskPage struct {
	*tview.Table
	*AppUI

	header []model.TaskHeader

	styles *Styles
	data   model.DataProducer
}

func NewTaskPage(app *AppUI, data model.DataProducer) *TaskPage {
	return &TaskPage{
		Table:  tview.NewTable(),
		AppUI:  app,
		styles: &PresetStyles,
		data:   data,
	}
}

func (t *TaskPage) Init() {
	t.Table.SetSelectable(true, false)

	t.header = []model.TaskHeader{
		{HeaderID},
		{HeaderDomain},
		{HeaderName},
		{HeaderStatus},
		{HeaderAge},
		{HeaderURL},
	}
}

func (t *TaskPage) Name() string {
	return TaskPageName
}

type Te struct {
	text  string
	color tcell.Color
}

func (t *TaskPage) AddRow(i int, task *model.TaskRow) {
	cID := &tview.TableCell{
		Text: strconv.FormatUint(task.ID, 10),
	}
	cID.SetReference(task.ID)
	t.SetCell(i, 1, cID)

	t.SetCell(i, 1, &tview.TableCell{
		Text: task.Domain,
	})
	t.SetCell(i, 2, &tview.TableCell{
		Text: task.Name,
	})
	t.SetCell(i, 3, &tview.TableCell{
		Text: task.State,
	})
	t.SetCell(i, 4, &tview.TableCell{
		Text: task.Age,
	})
	t.SetCell(i, 5, &tview.TableCell{
		Text: task.URL,
	})
}

func (t *TaskPage) Start() {
	t.Watch(context.Background())
}

func (t *TaskPage) Watch(ctx context.Context) {
	go t.update(ctx)
}

func (t *TaskPage) update(ctx context.Context) {
	RefreshRate := 100 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(RefreshRate):
			RefreshRate = 2 * time.Second
			t.AppUI.app.QueueUpdateDraw(func() {
				_ = t.refresh(ctx)
			})

		}
	}
}

func (t *TaskPage) refresh(ctx context.Context) error {
	rows := t.data.GetRows()
	if len(rows) > 0 {
		_, ok := rows[0].(*model.TaskRow)
		if !ok {
			return fmt.Errorf("expecting a meta table but got %T", rows[0])
		}
	}

	t.AddHeader()

	for i, row := range rows {
		t.AddRow(i+1, row.(*model.TaskRow))
	}
	r, _ := t.GetSelection()
	t.Select(r, 0)
	return nil
}

func (t *TaskPage) AddHeader() {
	_ = t.styles.Table.Header
	for i, h := range t.header {
		c := &tview.TableCell{
			Text:        h.Name,
			Transparent: true,
			Expansion:   1,
		}
		t.SetCell(0, i, c)
	}
}
