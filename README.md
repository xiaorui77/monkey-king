# monkey-king

Monkey King. 一个基于golang的爬虫软件.

## 使用

```golang
// 参照main中的示例编写OnHtml

// 浏览页面
collector.OnHTML(pageRe, func (ele *engine.HTMLElement) {
_ = ele.Request.Visit(ele.Attr[0].Val)
})

// 保存文件
collector.OnHTML(girlRe, func (e *engine.HTMLElement) {
name := e.GetText("body > div:nth-child(6) > div > h1", "girl-"+string(rand.Int31n(1000)))
file := fmt.Sprintf("%v-%03d", name, e.Index)
_ = e.Request.Download(file, fmt.Sprintf("%v/%v", basePath, name), e.Attr[0].Val)
})
```

```bash
# 快捷键
":": 打开命令模式, 取值: tasks, logs, 分别可以查看任务队列和日志
"/": 打开输入默认, 输入url即可开始爬取
```
