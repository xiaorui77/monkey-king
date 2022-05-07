package schedule

import (
	"context"
	"encoding/json"
	"github.com/xiaorui77/goutils/logx"
	timeutil "github.com/xiaorui77/goutils/time"
	"github.com/xiaorui77/monker-king/internal/engine/schedule/task"
	"github.com/xiaorui77/monker-king/internal/utils/fileutil"
	"gorm.io/gorm"
	"sync"
	"time"
)

// Browser 在同一个域名下的调度器
// 1. 处理同一个Domain下的优先级关系
// 2. 管理cookie等
type Browser struct {
	scheduler *Scheduler

	mu    sync.Mutex
	wg    sync.WaitGroup
	numCh chan int

	domain     string
	processNum int
	processes  []*Process

	MaxDepth int        // 最大层级, 包括下一页等
	taskList *task.List // 存储结构
}

func NewBrowser(s *Scheduler, domain string) *Browser {
	return &Browser{
		scheduler: s,
		domain:    domain,
		processes: make([]*Process, 0, 5),

		taskList: task.NewTaskList(),
		MaxDepth: MaxDepth,
	}
}

// schedule all tasks by multi-thread.
func (b *Browser) boot(ctx context.Context) {
	logx.Debugf("[scheduler] The Browser[%s] boot, processNum: %d", b.domain, Parallelism)
	b.setProcess(ctx, Parallelism)

	for {
		select {
		case <-ctx.Done():
			logx.Debugf("[scheduler] The Browser[%s] ctx.done, waiting all process stop", b.domain)
			b.wg.Wait()
			logx.Debugf("[scheduler] The Browser[%s] all process has been stopped", b.domain)
			b.close()
			logx.Infof("[scheduler] The Browser[%s] has been stopped", b.domain)
			return
		case num := <-b.numCh:
			b.setProcess(ctx, num)
		case <-time.Tick(time.Second * 20):
			logx.Debugf("[scheduler] The Browser[%s] run retryFailed", b.domain)
			b.retryFailed()
		}
	}
}

func (b *Browser) setProcess(ctx context.Context, num int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	logx.Infof("[scheduler] Browser[%s] set processNum: %d to %d", b.domain, b.processNum, num)
	if b.processNum < num {
		for index := b.processNum; index < num; index++ {
			cancelledCtx, cancel := context.WithCancel(ctx)
			p := &Process{
				browser: b, index: index, cancelFn: cancel,
				logger: logx.WithFields(logx.Fields{"browser": b.domain, "process": index}),
			}
			b.processes = append(b.processes, p)
			b.wg.Add(1)
			go func(index int) {
				defer func() {
					b.mu.Lock()
					defer b.mu.Unlock()
					b.wg.Done()
					b.processNum--
				}()
				p.run(cancelledCtx)
			}(index)
		}
	} else if b.processNum > num {
		// 调用cancel函数结束running
		for index := b.processNum - 1; index >= num; index-- {
			b.processes[index].cancelFn()
		}
	}
	b.processNum = num
}

func (b *Browser) SetProcess(num int) {
	b.numCh <- num
}

func (b *Browser) recordErr(t *task.Task, code int, msg string) {
	t.SetState(task.StateFailed)
	t.RecordErr(code, msg)
	if err := b.scheduler.store.GetDB().Save(t).UpdateColumn("err_num", len(t.ErrDetails)).Error; err != nil {
		logx.Errorf("[storage] update task[08x] state error: %v", t.ID, err)
	}
}

func (b *Browser) recordStart(t *task.Task) {
	t.SetState(task.StateRunning)
	if err := b.scheduler.store.GetDB().Save(t).Error; err != nil {
		logx.Errorf("[storage] update task[%08x] error: %v", t.ID, err)
	}
}

func (b *Browser) recordSuccess(t *task.Task) {
	t.SetState(task.StateSuccessful)
	if err := b.scheduler.store.GetDB().Save(t).Error; err != nil {
		logx.Errorf("[storage] update task[%08x] error: %v", t.ID, err)
	}
}

func (b *Browser) timeout(t *task.Task) (tt time.Duration) {
	defer func() {
		// defer + func() {} 的形式是可以将返回值传进来的, 如果是defer直接+t.SetMeta(), 则tt=0
		t.SetMeta(task.MetaTimeout, int64(tt.Seconds()))
	}()
	if len(t.ErrDetails) == 0 {
		return DefaultTimeout
	}
	// 基于上次reader的情况计算超时时间
	lastTimeout, ltOk := t.Meta[task.MetaTimeout].(int64)
	reader, rOk := t.Meta[task.MetaReader].(*fileutil.VisualReader)
	if ltOk && rOk && lastTimeout > 0 && reader.Cur > 0 && reader.Total > 0 {
		timeout := lastTimeout * reader.Total / reader.Cur * int64(len(t.ErrDetails)+1)
		return timeutil.Min(DefaultTimeout+time.Second*time.Duration(timeout), MaxTimeout)
	}
	return timeutil.Min(DefaultTimeout+time.Second*45*time.Duration(len(t.ErrDetails)), MaxTimeout)
}

func (b *Browser) close() {
	// todo: close all task queue of the domain

	// 自我清理
	delete(b.scheduler.browsers, b.domain)
	b.scheduler = nil
}

func (b *Browser) next() *task.Task {
	b.mu.Lock()
	defer b.mu.Unlock()

	t := b.taskList.Next()
	if t == nil {
		return nil
	}
	if err := b.scheduler.store.GetDB().Model(t).UpdateColumn("state", t.State).Error; err != nil {
		logx.Errorf("[storage] update task[%08x] error: %v", t.ID, err)
	}
	return t
}

func (b *Browser) push(t *task.Task) {
	if t == nil || t.Depth > b.MaxDepth {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if t.Parent != nil {
		t.Parent.Push(t)
	} else {
		b.taskList.Push(t)
	}
	// 持久化
	if err := b.scheduler.store.GetDB().Create(t).Error; err != nil {
		logx.Errorf("[storage] save task[%08x] to db error: %v", t.ID, err)
	}
}

// todo: 需要替换
func (b *Browser) delete(id uint64) *task.Task {
	return nil
}

// todo: 需要替换
func (b *Browser) query(name string) *task.Task {
	return nil
}

func (b *Browser) retryFailed() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if t := b.taskList.RetryFailed(); t != nil {
		if err := b.scheduler.store.GetDB().Session(&gorm.Session{FullSaveAssociations: true}).Updates(t).Error; err != nil {
			logx.Errorf("[storage] update tasks state error: %v", err)
		}
	}
}

func (b *Browser) list() []*task.Task {
	return b.taskList.ListAll()
}

func (b *Browser) tree() *Browser {
	return b
}

func (b *Browser) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id         uint64     `json:"id"`
		Name       string     `json:"name"`
		ProcessNum int        `json:"processNum"`
		Children   *task.List `json:"children"`
	}{Id: 0, Name: b.domain, ProcessNum: b.processNum, Children: b.taskList})
}
