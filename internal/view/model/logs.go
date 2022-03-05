package model

type LogItem struct {
	Bytes []byte
}

type LogChan chan *LogItem

type LogsBuffer struct {
	LogChan
}

func NewLogsBuffer() *LogsBuffer {
	return &LogsBuffer{
		LogChan: make(LogChan, 100),
	}
}

func (l *LogsBuffer) GetLogChan() LogChan {
	return make(LogChan)
}

// Write implement io Writer interface
func (l *LogsBuffer) Write(p []byte) (int, error) {
	l.LogChan <- &LogItem{Bytes: p}
	return len(p), nil
}
