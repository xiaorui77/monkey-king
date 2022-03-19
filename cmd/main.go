package main

import (
	"fmt"
	"github.com/yougtao/monker-king/internal/manager"
	"math/rand"

	"github.com/yougtao/goutils/logx"
	"github.com/yougtao/goutils/wait"
	"github.com/yougtao/monker-king/internal/config"
	"github.com/yougtao/monker-king/internal/engine"
)

// girl
var (
	girlRe   = `body > div:nth-child(6) > div > div.pic > img`
	pageRe   = `body > div:nth-child(6) > div > div.row.col6.clearfix > dl > dt > a`
	pagingRe = `body > div:nth-child(8) > div > div.pc_pagination > a:nth-child(11)`
)

var basePath = "./data"

//var basePath = "D:\\tmp\\226g.net"

func main() {
	_, stopCtx := wait.SetupStopSignal()

	// option
	logx.Init(logx.OptLevel("debug"), logx.OptReportCaller(true))

	conf := config.InitConfig()
	collector, err := engine.NewCollector(conf)
	if err != nil {
		logx.Fatalf("[engine] create collector failed: %v", err)
		return
	}

	// 单页
	collector.OnHTMLAny(girlRe, func(e *engine.HTMLElement) {
		name := e.GetText("body > div:nth-child(6) > div > h1", "girl-"+string(rand.Int31n(1000)))
		file := fmt.Sprintf("%v-%03d", name, e.Index)
		path := fmt.Sprintf("%v/%v", basePath, name)
		_ = collector.Download(file, path, e.Attr[0].Val)
	})

	// 列表跳转到page
	collector.OnHTMLAny(pageRe, func(ele *engine.HTMLElement) {
		_ = collector.Visit(ele.Attr[0].Val)
	})

	// 分页
	collector.OnHTMLAny(pagingRe, func(ele *engine.HTMLElement) {
		_ = collector.Visit(ele.Attr[0].Val)
	})

	// begin
	//_ = collector.Visit("https://www.228n.net/pic/toupai/")

	// ui
	// ui := view.NewUI(collector)
	// ui.Init()
	// logx.SetOutput(ui.GetLogsWriter())
	// go ui.Run(stopCtx)

	// manager
	go manager.NewManager(collector).Run(stopCtx)

	collector.Run(stopCtx)
}
