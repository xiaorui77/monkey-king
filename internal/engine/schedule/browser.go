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
	schedule *Scheduler
	domain   string
	normal   *TaskQueue

	// 最大层级, 包括下一页等
	MaxDepth int
}

func NewDomainBrowser(s *Scheduler, host string) *DomainBrowser {
	return &DomainBrowser{
		schedule: s,
		domain:   host,
		normal:   NewTaskQueue(),

		MaxDepth: 6,
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
	logx.Infof("[scheduler] The Browser of domain[%s] has been closed", d.domain)
}

func (d *DomainBrowser) process(ctx context.Context, wg *sync.WaitGroup, index int) {
	for {
		select {
		case <-ctx.Done():
			logx.Infof("[scheduler] [process-%d] The Process[%s-%d] will stop", index, d.domain, index)
			wg.Done()
			return
		default:
			t := d.next()
			if t == nil {
				time.Sleep(2 * time.Second)
				continue
			}
			logx.Infof("[scheduler] [process-%d] Task[%x] begin run, request url: %s", index, t.ID, t.Url)
			t.SetState(task.StateRunning)
			req, resp, err := d.schedule.download.Get(t)
			if err != nil {
				logx.Errorf("[scheduler] [process-%d] Task[%x] request(GET) fail: %v", index, t.ID, err)
				t.RecordErr(err.ErrCode(), err.Error())
			} else {
				logx.Infof("[scheduler] [process-%d] Task[%x] request.Do finish, will handle Response", index, t.ID)
				t.HandleOnResponse(req, resp)
				logx.Infof("[scheduler] [process-%d] Task[%x] success", index, t.ID)
			}
		}
		time.Sleep(time.Second * 3)
	}
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
		if tq.tasks[j].State == task.StateInit || tq.tasks[j].State == task.StateFail {
			tq.offset = j + 1
			tq.tasks[j].SetState(task.StateUnknown)
			return tq.tasks[j]
		}
	}
	return nil
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
