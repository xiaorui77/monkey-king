package main

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/yougtao/goutils/logx"
	"github.com/yougtao/goutils/wait"
	"github.com/yougtao/monker-king/internal/config"
	"github.com/yougtao/monker-king/internal/engine"
	"math/rand"
)

// girl
var (
	girlRe   = `body > div:nth-child(6) > div > div.pic > img`
	pageRe   = `body > div:nth-child(6) > div > div.row.col6.clearfix > dl > dt > a`
	pagingRe = `body > div:nth-child(8) > div > div.pc_pagination > a:nth-child(11)`
)

var basePath = "~/226g.net"

//var basePath = "D:\\tmp\\226g.net"

func main() {
	_, stopCtx := wait.SetupStopSignal()

	buf := &bytes.Buffer{}
	logx.SetLevel(logrus.DebugLevel)
	logx.SetOutput(buf)

	conf := config.InitConfig()
	collector, err := engine.NewCollector(conf)
	if err != nil {
		logx.Fatalf("[engine] create collector failed: %v", err)
		return
	}

	// 单页
	collector.OnHTML(girlRe, func(e *engine.HTMLElement) {
		name := e.GetText("body > div:nth-child(6) > div > h1", "girl-"+string(rand.Int31n(1000)))
		file := fmt.Sprintf("%v-%03d", name, e.Index)
		_ = e.Request.Download(file, fmt.Sprintf("%v/%v", basePath, name), e.Attr[0].Val)
	})

	// 列表跳转到page
	collector.OnHTML(pageRe, func(ele *engine.HTMLElement) {
		_ = ele.Request.Visit(ele.Attr[0].Val)
	})

	// 分页
	collector.OnHTML(pagingRe, func(ele *engine.HTMLElement) {
		_ = ele.Request.Visit(ele.Attr[0].Val)
	})

	// begin
	//_ = collector.Visit("https://www.228n.net/pic/toupai/")
	collector.Run(stopCtx)
}
