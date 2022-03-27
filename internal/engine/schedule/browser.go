package schedule

import (
	"context"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/goutils/wait"
	"github.com/xiaorui77/monker-king/internal/engine/task"
	"math/rand"
	"sync"
	"time"
)

// DomainBrowser 在同一个域名下的调度器
// 1. 处理同一个Domain下的优先级关系
// 2. 管理cookie等
type DomainBrowser struct {
	schedule *Scheduler
	domain   string
	normal   *TaskQueue
}

func NewDomainBrowser(s *Scheduler, host string) *DomainBrowser {
	return &DomainBrowser{
		schedule: s,
		domain:   host,
		normal:   NewTaskQueue(),
	}
}

func (d *DomainBrowser) push(task *task.Task) {
	d.normal.push(task)
}

// todo
func (d *DomainBrowser) next() *task.Task {
	return d.normal.next()
}

func (d *DomainBrowser) list() []*task.Task {
	res := make([]*task.Task, 0, len(d.normal.list()))
	res = append(res, d.normal.list()...)
	return res
}

// schedule all tasks by multi-thread.
func (d *DomainBrowser) begin(ctx context.Context) {
	logx.Infof("[schedule] The Browser of domain[%s] begin, process num: %d", d.domain, Parallelism)
	var wg sync.WaitGroup
	for i := 0; i < Parallelism; i++ {
		// 因为可能在创建ctx之前, 已经有任务被添加进来了
		wait.WaitUntil(func() bool { return ctx != nil })
		go d.process(ctx, &wg, i)
		wg.Add(1)
	}

	wg.Wait()
	d.close()
	logx.Infof("[schedule] The Browser of domain[%s] has been closed", d.domain)
}

func (d *DomainBrowser) process(ctx context.Context, wg *sync.WaitGroup, index int) {
	for {
		select {
		case <-ctx.Done():
			logx.Infof("[schedule] The process[%s-%d] will stop", d.domain, index)
			wg.Done()
			return
		default:
			t := d.next()
			if t == nil {
				time.Sleep(2 * time.Second)
				continue
			}
			if t.ID == 0 {
				t.ID = rand.Uint64()
			}
			logx.Infof("[schedule-%d] The task[%x] begin to run, url: %s", index, t.ID, t.Url)
			t.SetState(task.StateRunning)
			d.schedule.download.Get(t)
			time.Sleep(time.Second * 3)
			logx.Debugf("[schedule-%d] The task[%x] done.", index, t.ID)
		}
	}
}

func (d *DomainBrowser) reFail(_ *task.Task) {

}

func (d *DomainBrowser) close() {
	// todo: close all task queue of the domain
}

func (d *DomainBrowser) MarshalJSON() ([]byte, error) {
	return nil, nil
}

// TaskQueue 任务队列
type TaskQueue struct {
	sync.RWMutex

	tasks  []*task.Task
	offset int
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		tasks: []*task.Task{},
	}
}

func (tq *TaskQueue) push(task *task.Task) {
	tq.Lock()
	defer tq.Unlock()

	l := len(tq.tasks) - 1
	for i := range tq.tasks {
		j := l - i
		if tq.tasks[j].Priority <= task.Priority {
			tq.tasks = append(tq.tasks, nil)
			copy(tq.tasks[j+1:], tq.tasks[j:])
			tq.tasks[j] = task

			if j < tq.offset {
				tq.offset = j
			}
			return
		}
	}
	tq.tasks = append(tq.tasks, task)
}

func (tq *TaskQueue) next() *task.Task {
	tq.Lock()
	defer tq.Unlock()

	for i, t := range tq.tasks {
		if i >= tq.offset && t.State != task.StateSuccess {
			tq.offset = i + 1
			if tq.offset > len(tq.tasks) {
				tq.offset = 0
			}
			return t
		}
	}
	return nil
}

func (tq *TaskQueue) list() []*task.Task {
	return tq.tasks
}
