package task

import (
	"encoding/json"
	"github.com/xiaorui77/goutils/logx"
)

type TaskList struct {
	list   []*Task
	offset int
}

func NewTaskList() *TaskList {
	return &TaskList{list: make([]*Task, 0, 10)}
}

func (l *TaskList) Push(t *Task) {
	for i := len(l.list) - 1; i >= 0; i-- {
		if t.Priority <= l.list[i].Priority {
			l.list = append(l.list, nil)
			copy((l.list)[i+2:], (l.list)[i+1:])
			l.list[i+1] = t

			if l.offset > i+1 {
				l.offset = i + 1
			}
			return
		}
	}
	// 插入前部
	l.list = append([]*Task{t}, l.list...)
	l.offset = 0
}

func (l *TaskList) Next() *Task {
	first := true
	for i := l.offset; i < l.offset+len(l.list); i++ {
		j := i % len(l.list)
		if l.list[j].State == StateInit {
			l.list[j].SetState(StateScheduling)
			return l.list[j]
		} else if l.list[j].State == StateSuccessful {
			// 深度优先
			if n := l.list[j].nextSub(); n != nil {
				return n
			}
		} else if first && (l.list[j].State == StateSuccessfulAll || l.list[j].State == StateFailed) { // 暂时跳过failed还是等待failed
			first = false
			l.offset = j + 1
		}
	}
	return nil
}

func (l *TaskList) RetryFailed() {
	for _, t := range l.list {
		if t.State != StateFailed || len(t.ErrDetails) == 0 {
			// 非错误或者错误无详情时调过分析
			continue
		}
		if len(t.ErrDetails) > 7 {
			logx.Warnf("[browser] Task[%x] failure more than 7 times, will no longer try again", t.ID)
			continue // 超过7次不再重试
		}
		n := 0
		for i := len(t.ErrDetails) - 1; i >= 0; i-- {
			if t.ErrDetails[i].ErrCode == ErrHttpNotFount {
				n++
			} else {
				break
			}
		}
		if n < 2 {
			// 连续的NotFound错误小于2次才重试
			logx.Infof("[browser] Task[%x] can be retry, last err: %s", t.ID, t.ErrDetails[len(t.ErrDetails)-1].String())
			t.SetState(StateInit)
			if t.Parent != nil {
				t.Parent.Children.offset = 0
			}
		}
	}

	for _, t := range l.list {
		if t.Children != nil {
			t.Children.RetryFailed()
		}
	}
}

func (l *TaskList) isSuccessfulAll() bool {
	for _, t := range l.list {
		if t.IsSuccessful() == false {
			return false
		}
	}
	return true
}

func (l *TaskList) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.list)
}

func (l *TaskList) ListAll() []*Task {
	res := make([]*Task, 0, len(l.list))
	res = append(res, l.list...)

	for i := 0; i < len(res); i++ {
		if res[i].Children != nil {
			res = append(res, res[i].Children.list...)
		}
	}
	return res
}
