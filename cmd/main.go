package main

import (
	"context"
	"fmt"
	"github.com/xiaorui77/goutils/fileutils"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/goutils/logx/hooks"
	"github.com/xiaorui77/goutils/math"
	"github.com/xiaorui77/goutils/wait"
	"github.com/xiaorui77/monker-king/internal/config"
	"github.com/xiaorui77/monker-king/internal/engine/collector"
	"github.com/xiaorui77/monker-king/internal/engine/task"
	"github.com/xiaorui77/monker-king/internal/manager"
	"math/rand"
	"os/signal"
)

var (
	girlRe   = `body > div:nth-child(6) > div > div.pic img`
	pageRe   = `body > div:nth-child(6) > div > div.row.col6.clearfix > dl > dt > a`
	pagingRe = `body > div:nth-child(8) > div > div.pc_pagination > a:nth-last-child(2)`
)

var basePath = "./data"

func main() {
	stopCtx, _ := signal.NotifyContext(context.Background(), wait.ShutdownSignals...)

	// option
	logx.Init("monkey-king", logx.WithInstance("monkey-king-"+math.RandomStr(5, 36)),
		logx.WithLevel(logx.DebugLevel), logx.WithReportCaller(true),
		logx.WithHook(hooks.NewEsHook("http://192.168.17.1:9200")))

	conf := config.InitConfig()
	engine, err := collector.NewCollector(conf)
	if err != nil {
		logx.Fatalf("[engine] create collector failed: %v", err)
		return
	}

	// 每个单元下所有元素
	engine.OnHTMLAny(girlRe, func(t *task.Task, e *collector.HTMLElement) {
		name := e.GetText("body > div:nth-child(6) > div > h1", "girl-"+string(rand.Int31n(1000)))
		name = fileutils.WindowsName(name)
		file := fmt.Sprintf("%v-%03d", name, e.Index)
		path := fmt.Sprintf("%v/%v", basePath, name)
		_ = engine.Download(t, file, path, e.Attr[0].Val)
	})

	// 每页内所有单元
	engine.OnHTMLAny(pageRe, func(t *task.Task, ele *collector.HTMLElement) {
		_ = ele.Visit(ele.Attr[0].Val)
	})

	// 下个页
	engine.OnHTMLAny(pagingRe, func(t *task.Task, ele *collector.HTMLElement) {
		_ = ele.Visit(ele.Attr[0].Val)
	})

	// ui
	// ui := view.NewUI(collector)
	// ui.Init()
	// logx.SetOutput(ui.GetLogsWriter())
	// go ui.Run(stopCtx)

	// manager
	go manager.NewManager(engine).Run(stopCtx)

	engine.Run(stopCtx)
	logx.Infof("main has been exit")
}
