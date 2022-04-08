package schedule

import (
	"context"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/goutils/timeutils"
	"github.com/xiaorui77/monker-king/internal/engine/task"
	"github.com/xiaorui77/monker-king/internal/utils/fileutil"
	"sync"
	"time"
)

// Browser 在同一个域名下的调度器
// 1. 处理同一个Domain下的优先级关系
// 2. 管理cookie等
type Browser struct {
	scheduler *Scheduler
	domain    string
	normal    *TaskQueue

	// 最大层级, 包括下一页等
	MaxDepth int
}

func NewBrowser(s *Scheduler, host string) *Browser {
	return &Browser{
		scheduler: s,
		domain:    host,
		normal:    NewTaskQueue(),

		MaxDepth: MaxDepth,
	}
}

func (b *Browser) push(task *task.Task) {
	if task == nil || task.Depth > b.MaxDepth {
		return
	}
	b.normal.push(task)
}

func (b *Browser) delete(name string) *task.Task {
	return b.normal.delete(name)
}

func (b *Browser) query(name string) *task.Task {
	return b.normal.query(name)
}

// todo
func (b *Browser) next() *task.Task {
	return b.normal.next()
}

func (b *Browser) refresh() {
	b.normal.refresh()
}

func (b *Browser) list() []*task.Task {
	res := make([]*task.Task, 0, len(b.normal.list()))
	res = append(res, b.normal.list()...)
	return res
}

// schedule all tasks by multi-thread.
func (b *Browser) begin(ctx context.Context) {
	logx.Infof("[scheduler] The Browser[%s] begin, process num: %d", b.domain, Parallelism)
	var wg sync.WaitGroup
	for i := 0; i < Parallelism; i++ {
		index := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			logx.Infof("[scheduler] Browser[%s] Process[%d] start process", b.domain, index)
			for {
				select {
				case <-ctx.Done():
					logx.Infof("[scheduler] Browser[%s] Process[%d] will stop", b.domain, index)
					return
				default:
					b.process(ctx, index)
				}
				time.Sleep(time.Second * TaskInterval)
			}
		}()
	}

	wg.Wait()
	logx.Infof("[scheduler] The Browser[%s] will close", b.domain)
	b.close()
	logx.Infof("[scheduler] The Browser[%s] has been closed", b.domain)
}

func (b *Browser) process(ctx context.Context, index int) {
	t := b.next()
	if t == nil {
		logx.Debugf("[process-%d] no found tasks", index)
		b.refresh()
		time.Sleep(time.Second * 3)
		return
	}
	logx.Infof("[process-%d] Task[%x] begin run, request url: %s", index, t.ID, t.Url)
	t.RecordStart()

	// 设置超时并使用GET进行请求
	tCtx, cancelFunc := context.WithTimeout(ctx, b.timeout(t))
	defer cancelFunc()
	req, resp, err := b.scheduler.download.Get(tCtx, t)
	if err != nil {
		logx.Errorf("[process-%d] Task[%x] request(GET) fail: %v", index, t.ID, err)
		t.RecordErr(err.ErrCode(), err.Error())
		return
	}

	logx.Infof("[scheduler] [process-%d] Task[%x] request.Do finish, will handle OnResponse", index, t.ID)
	if err := t.HandleOnResponse(req, resp); err != nil {
		logx.Errorf("[process-%d] Task[%x] handle OnResponse failed: %v", index, t.ID, err)
		t.RecordErr(err.ErrCode(), err.Error())
		return
	}
	t.RecordSuccess()
	logx.Infof("[process-%d] Task[%x] run success", index, t.ID)
}

func (b *Browser) timeout(t *task.Task) (tt time.Duration) {
	defer func() {
		// defer + func() {} 的形式是可以将返回值传进来的, 如果是defer直接+t.SetMeta(), 则tt=0
		t.SetMeta("timeout", tt)
	}()
	if len(t.ErrDetails) == 0 {
		return DefaultTimeout
	}
	// 基于上次reader的情况计算超时时间
	lastTimeout, ltOk := t.Meta["timeout"].(time.Duration)
	reader, rOk := t.Meta["reader"].(*fileutil.VisualReader)
	if ltOk && rOk && lastTimeout > 0 && reader.Cur > 0 && reader.Total > 0 {
		dur := lastTimeout * time.Duration(reader.Total) / time.Duration(reader.Cur)
		return timeutils.Min(dur, MaxTimeout)
	}
	return timeutils.Min(DefaultTimeout+time.Second*45*time.Duration(len(t.ErrDetails)), MaxTimeout)
}

func (b *Browser) close() {
	// todo: close all task queue of the domain

	// 自我清理
	delete(b.scheduler.browsers, b.domain)
	b.normal = nil
	b.scheduler = nil
}

func (b *Browser) MarshalJSON() ([]byte, error) {
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
		if t.State == task.StateFail && len(t.ErrDetails) > 0 {
			if len(t.ErrDetails) > 7 {
				continue // 超过7次不再重试
			}
			n := 0
			for i := len(t.ErrDetails) - 1; i >= 0; i-- {
				if t.ErrDetails[i].ErrCode == task.ErrHttpNotFount {
					n++
				} else {
					break
				}
			}
			if n < 2 {
				// 连续的NotFound错误小于2次才重试
				logx.Infof("[browser] Task[%x] can be retry, last err: %v", t.ID, t.ErrDetails[len(t.ErrDetails)-1])
				t.SetState(task.StateInit)
			}
		}
	}
	tq.offset = 0
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
