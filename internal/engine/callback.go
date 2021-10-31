package engine

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

type HtmlCallback func(element *HTMLElement)
type HtmlCallbackContainer struct {
	Selector string
	fun      HtmlCallback
}

func (c *Collector) handleOnHtml(resp *Response) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body))
	if err != nil {
		logrus.Debugf("parse html to document failed: %v", err)
		return
	}
	for _, callback := range c.htmlCallbacks {
		index := 1
		doc.Find(callback.Selector).Each(func(_ int, selection *goquery.Selection) {
			for _, node := range selection.Nodes {
				e := NewHTMLElement(resp, doc, selection, node, index)
				index++
				callback.fun(e)
			}
		})
	}
}
