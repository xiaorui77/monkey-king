package logx_hooks

import (
	"fmt"
	"github.com/xiaorui77/goutils/logx"
	"strings"
)

type format struct {
	logger *logx.LogX
}

func NewPostFormat() *format {
	return &format{}
}

func (f *format) SetLogger(logger *logx.LogX) {
	f.logger = logger
}

func (f *format) Fire(entry *logx.Entry) error {
	if entry == nil || entry.Fields == nil {
		return nil
	}
	if strings.HasPrefix(entry.Message, "[") {
		// pass
		return nil
	}
	prefix := ""
	if catalog, ok := entry.Fields["catalog"]; ok {
		prefix += fmt.Sprintf("[%v] ", catalog)
	}
	if browser, ok := entry.Fields["browser"]; ok {
		prefix += fmt.Sprintf("Browser[%s]", browser)
	}
	if process, ok := entry.Fields["process"]; ok {
		prefix += fmt.Sprintf("[process-%d]", process)
	}
	if taskId, ok := entry.Fields["taskId"]; ok {
		prefix += fmt.Sprintf("Task[%08x]", taskId)
	}
	entry.Message = prefix + entry.Message
	return nil
}

func (f *format) Levels() []logx.Level {
	return []logx.Level{logx.DebugLevel, logx.InfoLevel, logx.WarnLevel, logx.ErrorLevel, logx.FatalLevel}
}
