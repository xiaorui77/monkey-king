package schedule

import (
	"context"
	"encoding/json"
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
	scheduler  *Scheduler
	domain     string
	processNum int

	mu            sync.Mutex
	wg            sync.WaitGroup
	processCancel []context.CancelFunc

	// 最大层级, 包括下一页等
	MaxDepth int
	taskTree *task.Tree

	// 新存储结构
	tasks *task.TaskList
}

func NewBrowser(s *Scheduler, domain string) *Browser {
	return &Browser{
		scheduler:     s,
		domain:        domain,
		processCancel: make([]context.CancelFunc, 0),

		tasks:    task.NewTaskList(),
		MaxDepth: MaxDepth,
	}
}

// schedule all tasks by multi-thread.
func (b *Browser) boot(ctx context.Context) {
	logx.Debugf("[scheduler] The Browser[%s] boot, processNum: %d", b.domain, Parallelism)
	b.setProcess(ctx, Parallelism)

	// 等待运行结束
	b.wg.Wait()
	logx.Infof("[scheduler] Browser[%s] will close", b.domain)
	b.close()
	logx.Infof("[scheduler] Browser[%s] has been stopped", b.domain)
}

func (b *Browser) setProcess(ctx context.Context, num int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	logx.Infof("[scheduler] Browser[%s] set processNum: %d to %d", b.domain, b.processNum, num)
	if b.processNum < num {
		for index := b.processNum; index < num; index++ {
			cancelledCtx, cancel := context.WithCancel(ctx)
			b.processCancel = append(b.processCancel, cancel)
			b.wg.Add(1)
			go func(index int) {
				defer b.wg.Done()
				defer func() {
					if err := recover(); err != nil {
						logx.Errorf("[scheduler] Browser[%s] process[%d] panic: %v", b.domain, index, err)
					}
				}()
				b.running(cancelledCtx, index)
			}(index)
		}
	} else if b.processNum > num {
		// 调用cancel函数结束running
		for index := b.processNum - 1; index >= num; index-- {
			b.processCancel[index]()
		}
	}
	b.processNum = num
}

func (b *Browser) running(ctx context.Context, index int) {
	logx.Infof("[scheduler] Browser[%s] Process[%d] start running", b.domain, index)
	for {
		select {
		case <-ctx.Done():
			logx.Infof("[scheduler] Browser[%s] Process[%d] has been stopped", b.domain, index)
			return
		default:
			b.process(ctx, index)
		}
		time.Sleep(time.Second * TaskInterval)
	}
}

func (b *Browser) process(ctx context.Context, index int) {
	t := b.next()
	if t == nil {
		logx.Debugf("[process-%d] no found tasks", index)
		b.retryFailed()
		return
	}
	logx.Infof("[process-%d] Task[%x] begin run, request url: %s", index, t.ID, t.Url)
	t.RecordStart()

	// 设置超时并使用GET进行请求
	tCtx, cancelFunc := context.WithTimeout(ctx, b.timeout(t))
	defer cancelFunc()
	resp, err := b.scheduler.download.Get(tCtx, t)
	if err != nil {
		logx.Errorf("[process-%d] Task[%x] run fail, request(GET) fail: %v", index, t.ID, err)
		t.RecordErr(err.ErrCode(), err.Error())
		return
	}

	logx.Infof("[scheduler] [process-%d] Task[%x] request.Do finish, will handle Callbacks", index, t.ID)
	if err := b.scheduler.parsing.HandleOnResponse(resp); err != nil {
		logx.Errorf("[process-%d] Task[%x] run fail, handle ResponseCallback failed: %v", index, t.ID, err)
		t.RecordErr(err.ErrCode(), err.Error())
		return
	}
	if err := t.Callback(t, resp); err != nil {
		logx.Errorf("[process-%d] Task[%x] handle task.Callback failed: %v", index, t.ID, err)
		t.RecordErr(task.ErrCallbackTask, err.Error())
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
	b.scheduler = nil
}

func (b *Browser) next() *task.Task {
	return b.tasks.Next()
}

func (b *Browser) push(t *task.Task) {
	if t == nil || t.Depth > b.MaxDepth {
		return
	}
	if t.Parent != nil {
		t.Parent.Push(t)
	} else {
		b.tasks.Push(t)
	}
}

// todo: 需要替换
func (b *Browser) delete(id uint64) *task.Task {
	// return b.taskQueue.Delete(t.Name)
	return b.taskTree.Delete(id)
}

// todo: 需要替换
func (b *Browser) query(name string) *task.Task {
	//b.taskQueue.Query(name)
	return b.taskTree.Query(name)
}

func (b *Browser) retryFailed() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.tasks.RetryFailed()
}

func (b *Browser) list() []*task.Task {
	return b.tasks.ListAll()
}

func (b *Browser) tree() *Browser {
	return b
}

func (b *Browser) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id         uint64       `json:"id"`
		Name       string       `json:"name"`
		ProcessNum int          `json:"processNum"`
		Children   []*task.Task `json:"children"`
	}{Id: 0, Name: b.domain, ProcessNum: b.processNum})
}
