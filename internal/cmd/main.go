package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/yougtao/monker-king/internal/engine"
	"math/rand"
)

// girl
var girlRe = `body > div:nth-child(6) > div > div.pic > img`
var pageRe = `body > div:nth-child(6) > div > div.row.col6.clearfix > dl > dt > a`
var pagingRe = `body > div:nth-child(8) > div > div.pc_pagination > a:nth-child(11)`

var basePath = "D:\\tmp\\226g.net"

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	collector := engine.NewCollector()

	// 单页
	collector.OnHTML(girlRe, func(e *engine.HTMLElement) {
		name := e.GetText("body > div:nth-child(6) > div > h1", "girl-"+string(rand.Int31n(1000)))
		file := fmt.Sprintf("%v-%03d", name, e.Index)
		path := fmt.Sprintf("%v/%v", basePath, name)
		_ = e.Request.Download(file, path, e.Attr[0].Val)
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
	collector.Visit("https://www.226g.net/pic/toupai/")
	collector.Run(context.TODO())
}
