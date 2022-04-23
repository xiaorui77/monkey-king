package task

import (
	"encoding/json"
	"github.com/xiaorui77/goutils/logx"
)

type List struct {
	Tasks  []*Task `gorm:"foreignKey:ParentId"`
	offset int     `gorm:"-"`
}

func NewTaskList() *List {
	return &List{Tasks: make([]*Task, 0, 10)}
}

func (l *List) Push(t *Task) {
	for i := len(l.Tasks) - 1; i >= 0; i-- {
		if t.Priority <= l.Tasks[i].Priority {
			l.Tasks = append(l.Tasks, nil)
			copy((l.Tasks)[i+2:], (l.Tasks)[i+1:])
			l.Tasks[i+1] = t

			if l.offset > i+1 {
				l.offset = i + 1
			}
			return
		}
	}
	// 插入前部
	l.Tasks = append([]*Task{t}, l.Tasks...)
	l.offset = 0
}

func (l *List) Next() *Task {
	first := true
	for i := l.offset; i < l.offset+len(l.Tasks); i++ {
		j := i % len(l.Tasks)
		if l.Tasks[j].State == StateInit {
			l.Tasks[j].SetState(StateScheduling)
			return l.Tasks[j]
		} else if l.Tasks[j].State == StateSuccessful {
			// 深度优先
			if n := l.Tasks[j].nextSub(); n != nil {
				return n
			}
		} else if first && (l.Tasks[j].State == StateSuccessfulAll || l.Tasks[j].State == StateFailed) { // 暂时跳过failed还是等待failed
			first = false
			l.offset = j + 1
		}
	}
	return nil
}

func (l *List) RetryFailed() {
	for _, t := range l.Tasks {
		if t.State != StateFailed || len(t.ErrDetails) == 0 {
			// 非错误或者错误无详情时调过分析
			continue
		}
		if len(t.ErrDetails) > 5 {
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

	for _, t := range l.Tasks {
		if t.Children != nil {
			t.Children.RetryFailed()
		}
	}
}

func (l *List) isSuccessfulAll() bool {
	for _, t := range l.Tasks {
		if t.IsSuccessful() == false {
			return false
		}
	}
	return true
}

func (l *List) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Tasks)
}

func (l *List) ListAll() []*Task {
	res := make([]*Task, 0, len(l.Tasks))
	res = append(res, l.Tasks...)

	for i := 0; i < len(res); i++ {
		if res[i].Children != nil {
			res = append(res, res[i].Children.Tasks...)
		}
	}
	return res
}
