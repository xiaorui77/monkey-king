package view

import (
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/xiaorui77/monker-king/internal/view/model"
	"io"
	"sync"
	"time"
)

const spacer = "     "

type LogsPage struct {
	*tview.Flex
	app *AppUI

	logsBuffer *model.LogsBuffer
	ansiWriter io.Writer

	indicator *tview.TextView
	logs      *tview.TextView

	autoscroll bool
	textWrap   bool

	cancelFn context.CancelFunc
	mx       sync.RWMutex
}

func NewLogsPage(app *AppUI) *LogsPage {
	return &LogsPage{
		Flex:       tview.NewFlex().SetDirection(tview.FlexRow),
		app:        app,
		logsBuffer: model.NewLogsBuffer(),

		autoscroll: true,
		textWrap:   false,
	}
}

func (l *LogsPage) Init() {
	l.SetBorder(true)
	l.SetDirection(tview.FlexRow)
	l.SetBorderPadding(0, 0, 1, 1)
	l.SetBackgroundColor(tcell.ColorDefault)

	l.indicator = tview.NewTextView()
	l.indicator.SetDynamicColors(true)
	l.indicator.SetTextColor(tcell.ColorWhite)
	l.indicator.SetBackgroundColor(tcell.ColorBlue)
	l.indicator.SetTextAlign(tview.AlignCenter)
	l.updateIndicator()
	l.Flex.AddItem(l.indicator, 1, 1, false)

	l.logs = tview.NewTextView()
	l.Flex.AddItem(l.logs, 0, 1, true)
	l.logs.SetScrollable(true).SetWrap(true).SetRegions(true)
	l.logs.SetDynamicColors(true)
	l.logs.SetBackgroundColor(tcell.ColorDefault)
	l.logs.SetText("[orange::d]" + "Waiting for logs...\n")
	l.logs.SetMaxLines(1000)

	l.ansiWriter = tview.ANSIWriter(l.logs)

	l.updateTitle()
}

func (l *LogsPage) Name() string {
	return LogsPageName
}

func (l *LogsPage) Start() {
	ctx := context.Background()
	ctx, l.cancelFn = context.WithCancel(ctx)

	go l.update(ctx)
}

func (l *LogsPage) update(ctx context.Context) {
	for {
		select {
		case item, ok := <-l.logsBuffer.LogChan:
			if !ok {
				return
			}
			_, _ = l.ansiWriter.Write(item.Bytes)

		case <-ctx.Done():
			return
		case <-time.After(time.Second * 3):
			l.LogChanged()
		}
	}
}

func (l *LogsPage) LogChanged() {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.app.app.QueueUpdateDraw(func() {
	})
}

func (l *LogsPage) refresh() {

}

func (l *LogsPage) Stop() {
	if l.cancelFn != nil {
		l.cancelFn()
		l.cancelFn = nil
	}
}

func (l *LogsPage) GetModel() *model.LogsBuffer {
	return l.logsBuffer
}

func (l *LogsPage) updateTitle() {
	l.Flex.SetTitle(fmt.Sprintf(" %s ", LogsPageName))
}

func (l *LogsPage) updateIndicator() {
	var text []byte
	if l.autoscroll {
		text = append(text, "[::b]Autoscroll:[limegreen::b]On[-::]"+spacer...)
	} else {
		text = append(text, "[::b]Autoscroll:[gray::d]Off[-::]"+spacer...)
	}

	if l.textWrap {
		text = append(text, "[::b]Wrap:[limegreen::b]On[-::]"...)
	} else {
		text = append(text, "[::b]Wrap:[gray::d]Off[-::]"...)
	}
	l.indicator.SetText(string(text))
}
