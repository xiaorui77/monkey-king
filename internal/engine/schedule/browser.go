package schedule

import (
	"context"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/goutils/wait"
	"github.com/xiaorui77/monker-king/internal/engine/task"
	"sync"
	"time"
)

// DomainBrowser 在同一个域名下的调度器
// 1. 处理同一个Domain下的优先级关系
// 2. 管理cookie等
type DomainBrowser struct {
	scheduler *Scheduler
	domain    string
	normal    *TaskQueue

	// 最大层级, 包括下一页等
	MaxDepth int
}

func NewDomainBrowser(s *Scheduler, host string) *DomainBrowser {
	return &DomainBrowser{
		scheduler: s,
		domain:    host,
		normal:    NewTaskQueue(),

		MaxDepth: MaxDepth,
	}
}

func (d *DomainBrowser) push(task *task.Task) {
	if task == nil || task.Depth > d.MaxDepth {
		return
	}
	d.normal.push(task)
}

func (d *DomainBrowser) delete(name string) *task.Task {
	return d.normal.delete(name)
}

func (d *DomainBrowser) query(name string) *task.Task {
	return d.normal.query(name)
}

// todo
func (d *DomainBrowser) next() *task.Task {
	return d.normal.next()
}

func (d *DomainBrowser) refresh() {
	d.normal.refresh()
}

func (d *DomainBrowser) list() []*task.Task {
	res := make([]*task.Task, 0, len(d.normal.list()))
	res = append(res, d.normal.list()...)
	return res
}

// schedule all tasks by multi-thread.
func (d *DomainBrowser) begin(ctx context.Context) {
	logx.Infof("[scheduler] The Browser of domain[%s] begin, process num: %d", d.domain, Parallelism)
	var wg sync.WaitGroup
	for index := 0; index < Parallelism; index++ {
		// 因为可能在创建ctx之前, 已经有任务被添加进来了
		wait.WaitUntil(func() bool { return ctx != nil })
		go func() {
			for {
				select {
				case <-ctx.Done():
					logx.Infof("[scheduler] [process-%d] The Process[%s-%d] will stop", index, d.domain, index)
					wg.Done()
					return
				default:
					d.process(index)
				}
				time.Sleep(time.Second * 3)
			}
		}()
		wg.Add(1)
	}

	wg.Wait()
	d.close()
	logx.Infof("[scheduler] The Browser of domain[%s] has been closed", d.domain)
}

func (d *DomainBrowser) process(index int) {
	t := d.next()
	if t == nil {
		d.refresh()
		logx.Debugf("[scheduler] [process-%d] no tasks", index)
		time.Sleep(time.Second * 3)
		return
	}
	logx.Infof("[scheduler] [process-%d] Task[%x] begin run, request url: %s", index, t.ID, t.Url)
	t.SetState(task.StateRunning)
	req, resp, err := d.scheduler.download.Get(t)
	if err != nil {
		logx.Errorf("[scheduler] [process-%d] Task[%x] request(GET) fail: %v", index, t.ID, err)
		t.RecordErr(err.ErrCode(), err.Error())
		return
	}

	logx.Infof("[scheduler] [process-%d] Task[%x] request.Do finish, will handle Response", index, t.ID)
	if err := t.HandleOnResponse(req, resp); err != nil {
		logx.Errorf("[scheduler] Task[%x] failed: %v", t.ID, err)
		t.RecordErr(err.ErrCode(), err.Error())
		return
	}
	t.SetState(task.StateSuccess)
	logx.Infof("[scheduler] [process-%d] Task[%x] run success", index, t.ID)
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

func (tq *TaskQueue) push(t *task.Task) {
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
	tq.tasks = append([]*task.Task{t}, tq.tasks...)
	tq.offset = 0
}

func (tq *TaskQueue) next() *task.Task {
	tq.Lock()
	defer tq.Unlock()

	for i := 0; i < len(tq.tasks); i++ {
		j := (tq.offset + i) % len(tq.tasks)
		if tq.tasks[j].State == task.StateInit {
			tq.offset = j + 1
			tq.tasks[j].SetState(task.StateUnknown)
			return tq.tasks[j]
		}
	}
	return nil
}

// 分析fail状态的task, 转为init
func (tq *TaskQueue) refresh() {
	tq.Lock()
	defer tq.Unlock()

	for _, t := range tq.tasks {
		if t.State == task.StateFail {
			if len(t.ErrDetails) > 0 {
				n := 0
				for _, detail := range t.ErrDetails {
					if detail.ErrCode == task.ErrHttpNotFount {
						n++
					} else {
						break
					}
				}
				if n > 2 {
					t.SetState(task.StateInit)
				}
				continue
			}
			t.SetState(task.StateInit)
		}
	}
}

func (tq *TaskQueue) query(name string) *task.Task {
	for _, t := range tq.tasks {
		if t.Name == name {
			return t
		}
	}
	return nil
}

func (tq *TaskQueue) delete(name string) *task.Task {
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

func (tq *TaskQueue) list() []*task.Task {
	return tq.tasks
}
