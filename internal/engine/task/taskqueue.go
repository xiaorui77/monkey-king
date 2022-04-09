package task

import (
	"github.com/xiaorui77/goutils/logx"
	"sync"
)

// Queue TaskQueue 任务队列
type Queue struct {
	sync.RWMutex

	tasks  []*Task
	offset int
}

func NewTaskQueue() *Queue {
	return &Queue{
		tasks: []*Task{},
	}
}

func (tq *Queue) Push(t *Task) {
	tq.Lock()
	defer tq.Unlock()

	for j := len(tq.tasks) - 1; j >= 0; j-- {
		if t.Priority <= tq.tasks[j].Priority {
			tq.tasks = append(tq.tasks, nil)
			copy(tq.tasks[j+2:], tq.tasks[j+1:])
			tq.tasks[j+1] = t

			if j+1 < tq.offset {
				tq.offset = j + 1
			}
			return
		}
	}
	// 插入前部
	tq.tasks = append([]*Task{t}, tq.tasks...)
	tq.offset = 0
}

func (tq *Queue) Next() *Task {
	tq.Lock()
	defer tq.Unlock()

	for i := 0; i < len(tq.tasks); i++ {
		j := (tq.offset + i) % len(tq.tasks)
		if tq.tasks[j].State == StateInit {
			tq.offset = j + 1
			tq.tasks[j].SetState(StateUnknown)
			return tq.tasks[j]
		}
	}
	return nil
}

// Refresh 分析fail状态的task, 转为init
func (tq *Queue) Refresh() {
	tq.Lock()
	defer tq.Unlock()

	for _, t := range tq.tasks {
		if t.State == StateFailed && len(t.ErrDetails) > 0 {
			if len(t.ErrDetails) > 7 {
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
				logx.Infof("[browser] Task[%x] can be retry, last err: %v", t.ID, t.ErrDetails[len(t.ErrDetails)-1])
				t.SetState(StateInit)
			}
		}
	}
	tq.offset = 0
}

func (tq *Queue) Query(name string) *Task {
	for _, t := range tq.tasks {
		if t.Name == name {
			return t
		}
	}
	return nil
}

func (tq *Queue) Delete(name string) *Task {
	tq.Lock()
	defer tq.Unlock()

	for i, t := range tq.tasks {
		if t.Name == name {
			tq.tasks = append(tq.tasks[:i], tq.tasks[i+1:]...)
			if i <= tq.offset {
				tq.offset--
			}
			return t
		}
	}
	return nil
}

func (tq *Queue) List() []*Task {
	return tq.tasks
}
