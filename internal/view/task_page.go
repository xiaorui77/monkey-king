package view

import (
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/yougtao/monker-king/internal/view/model"
	"strconv"
	"sync"
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

	actions  []ActionHandler
	cancelFn context.CancelFunc
	mx       sync.RWMutex
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
	t.SetFixed(1, 0)
	t.SetBorder(true)
	t.SetBorderPadding(0, 0, 1, 1)

	t.SetSelectable(true, false)
	t.SetSelectionChangedFunc(t.selectionChanged)
	t.SetBorderColor(tcell.ColorDefault)

	t.StylesChanged()

	t.header = []model.TaskHeader{
		{HeaderID},
		{HeaderDomain},
		{HeaderName},
		{HeaderStatus},
		{HeaderAge},
		{HeaderURL},
	}
}

func (t *TaskPage) addActions() {

}

func (t *TaskPage) handleRefresh() {

}

func (t *TaskPage) Name() string {
	return TaskPageName
}

type Te struct {
	text  string
	color tcell.Color
}

func (t *TaskPage) AddRow(i int, task *model.TaskRow) {
	color := t.getColor(task.State)

	cID := &tview.TableCell{
		Text:  strconv.FormatUint(task.ID, 16),
		Color: color,
	}
	cID.SetReference(task.ID)
	t.SetCell(i, 0, cID)

	t.SetCell(i, 1, &tview.TableCell{
		Text:  task.Domain,
		Color: color,
	})
	t.SetCell(i, 2, &tview.TableCell{
		Text:  task.Name,
		Color: color,
	})
	t.SetCell(i, 3, &tview.TableCell{
		Text:  task.State,
		Color: color,
	})
	t.SetCell(i, 4, &tview.TableCell{
		Text:  task.Age,
		Color: color,
	})
	t.SetCell(i, 5, &tview.TableCell{
		Text:  task.URL,
		Color: color,
	})
}

func (t *TaskPage) Start() {
	t.Stop()

	ctx := context.Background()
	ctx, t.cancelFn = context.WithCancel(ctx)

	go t.update(ctx)
}

func (t *TaskPage) Stop() {
	t.mx.Lock()
	if t.cancelFn != nil {
		t.cancelFn()
		t.cancelFn = nil
	}
	t.mx.Unlock()
}

func (t *TaskPage) update(ctx context.Context) {
	RefreshRate := 100 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(RefreshRate):
			RefreshRate = 2 * time.Second
			t.mx.RLock()
			fn := t.cancelFn
			t.mx.RUnlock()
			if fn == nil || !t.IsRunning() {
				continue
			}
			rows := t.data.GetRows()
			t.AppUI.app.QueueUpdateDraw(func() {
				_ = t.refresh(rows)
			})
		}
	}
}

func (t *TaskPage) refresh(rows []interface{}) error {
	if len(rows) > 0 {
		_, ok := rows[0].(*model.TaskRow)
		if !ok {
			return fmt.Errorf("expecting a meta table but got %T", rows[0])
		}
	}

	t.Table.Clear()

	// set title
	t.updateTitle(len(rows))

	// set table header
	t.AddHeader()

	for i, row := range rows {
		t.AddRow(i+1, row.(*model.TaskRow))
	}

	r, _ := t.Table.GetSelection()
	t.selectionChanged(r, 0)
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

func (t *TaskPage) updateTitle(total int) {
	t.Table.SetTitle(fmt.Sprintf(" %s(all) [%d] ", TaskPageName, total))
}

// ----------- styles -------------

func (t *TaskPage) selectionChanged(r, c int) {
	if r < 0 {
		return
	}
	if cell := t.GetCell(r, c); cell != nil {
		t.SetSelectedStyle(tcell.StyleDefault.
			Foreground(t.styles.Table.CursorFgColor.Color()).
			Background(cell.Color).
			Attributes(tcell.AttrBold))
	}
}

func (t *TaskPage) StylesChanged() {
}

func (t *TaskPage) getColor(state string) tcell.Color {
	switch state {
	case "init":
		return tcell.ColorOrange
	case "running":
		return tcell.ColorForestGreen
	case "Success":
		return tcell.ColorBlue
	case "fail":
		return tcell.ColorRed
	default:
		return tcell.ColorRed
	}
}
