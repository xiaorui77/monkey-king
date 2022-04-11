package task

import "encoding/json"

type TaskList struct {
	list   []*Task
	offset int
}

func newTaskList() *TaskList {
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
	for i := 0; i < len(l.list); i++ {
		j := (l.offset + i) % len(l.list)
		if l.list[j].State == StateInit {
			l.offset = j + 1
			l.list[j].SetState(StateScheduling)
			return l.list[j]
		}
	}
	for i := 0; i < len(l.list); i++ {
		if n := l.list[i].next(); n != nil {
			return n
		}
	}
	return nil
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
		if res[i].children != nil {
			res = append(res, res[i].children.list...)
		}
	}
	return res
}
