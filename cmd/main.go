package main

import (
	"fmt"
	"github.com/xiaorui77/goutils/logx/hooks"
	"github.com/xiaorui77/goutils/math"
	"github.com/xiaorui77/monker-king/internal/engine/task"
	"github.com/xiaorui77/monker-king/internal/manager"
	"math/rand"

	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/goutils/wait"
	"github.com/xiaorui77/monker-king/internal/config"
	"github.com/xiaorui77/monker-king/internal/engine"
)

var (
	girlRe   = `body > div:nth-child(6) > div > div.pic > img`
	pageRe   = `body > div:nth-child(6) > div > div.row.col6.clearfix > dl > dt > a`
	pagingRe = `body > div:nth-child(8) > div > div.pc_pagination > a:nth-child(11)`
)

var basePath = "./data"

func main() {
	_, stopCtx := wait.SetupStopSignal()

	// option
	logx.Init("monkey-king", logx.WithInstance("monkey-king-"+math.RandomStr(5, 36)),
		logx.WithLevel(logx.DebugLevel), logx.WithReportCaller(true),
		logx.WithHook(hooks.NewEsHook("http://192.168.43.104:9200")))

	conf := config.InitConfig()
	collector, err := engine.NewCollector(conf)
	if err != nil {
		logx.Fatalf("[engine] create collector failed: %v", err)
		return
	}

	// 单页
	collector.OnHTMLAny(girlRe, func(t *task.Task, e *engine.HTMLElement) {
		name := e.GetText("body > div:nth-child(6) > div > h1", "girl-"+string(rand.Int31n(1000)))
		file := fmt.Sprintf("%v-%03d", name, e.Index)
		path := fmt.Sprintf("%v/%v", basePath, name)
		_ = collector.Download(t, file, path, e.Attr[0].Val)
	})

	// 列表跳转到page
	collector.OnHTMLAny(pageRe, func(t *task.Task, ele *engine.HTMLElement) {
		_ = collector.Visit(t, ele.Attr[0].Val)
	})

	// 分页
	collector.OnHTMLAny(pagingRe, func(t *task.Task, ele *engine.HTMLElement) {
		_ = collector.Visit(t, ele.Attr[0].Val)
	})

	// ui
	// ui := view.NewUI(collector)
	// ui.Init()
	// logx.SetOutput(ui.GetLogsWriter())
	// go ui.Run(stopCtx)

	// manager
	go manager.NewManager(collector).Run(stopCtx)

	collector.Run(stopCtx)
}
