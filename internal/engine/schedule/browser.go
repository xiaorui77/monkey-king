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
	scheduler *Scheduler

	mu sync.Mutex
	wg sync.WaitGroup

	domain     string
	processNum int
	processes  []*process

	MaxDepth int            // 最大层级, 包括下一页等
	taskList *task.TaskList // 存储结构
}

func NewBrowser(s *Scheduler, domain string) *Browser {
	return &Browser{
		scheduler: s,
		domain:    domain,
		processes: make([]*process, 0, 5),

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
		case <-time.Tick(time.Second * 10):
			logx.Debugf("[scheduler] The Browser[%s] will run retryFailed", b.domain)
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
			p := &process{browser: b, index: index, cancelFn: cancel}
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

// 一个工作线程
type process struct {
	browser  *Browser
	index    int                // 计数
	cancelFn context.CancelFunc // 停止函数
}

func (p *process) run(ctx context.Context) {
	defer func() {
		defer p.cancelFn()
		logx.Errorf("[scheduler] Browser[%s] process[%d] has panic", p.browser.domain, p.index)
		if err := recover(); err != nil {
			logx.Errorf("[scheduler] Browser[%s] process[%d] panic, recover failed: %v", p.browser.domain, p.index, err)
		}
	}()

	logx.Infof("[scheduler] Browser[%s] Process[%d] has already started...", p.browser.domain, p.index)
	for {
		select {
		case <-ctx.Done():
			logx.Infof("[scheduler] Browser[%s] Process[%d] has been stopped", p.browser.domain, p.index)
			return
		default:
			p.process(ctx, p.index)
		}
		time.Sleep(time.Second * TaskInterval)
	}
}

// 工作过程
func (p *process) process(ctx context.Context, index int) {
	t := p.browser.next()
	if t == nil {
		logx.Debugf("[process-%d] no found tasks", index)
		return
	}
	logx.Infof("[process-%d] Task[%x] begin run, url: %s", index, t.ID, t.Url)
	t.RecordStart()

	// 设置超时并使用GET进行请求
	tCtx, cancelFunc := context.WithTimeout(ctx, p.browser.timeout(t))
	defer cancelFunc()
	resp, err := p.browser.scheduler.download.Get(tCtx, t)
	if err != nil {
		logx.Errorf("[process-%d] Task[%x] run fail, request(GET) fail: %v", index, t.ID, err)
		t.RecordErr(err.ErrCode(), err.Error())
		return
	}

	logx.Infof("[process-%d] Task[%x] request finish, will handle Callbacks", index, t.ID)
	if err := p.browser.scheduler.parsing.HandleOnResponse(resp); err != nil {
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
		return timeutils.Min(DefaultTimeout+dur, MaxTimeout)
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
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.taskList.Next()
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

	b.taskList.RetryFailed()
}

func (b *Browser) list() []*task.Task {
	return b.taskList.ListAll()
}

func (b *Browser) tree() *Browser {
	return b
}

func (b *Browser) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id         uint64         `json:"id"`
		Name       string         `json:"name"`
		ProcessNum int            `json:"processNum"`
		Children   *task.TaskList `json:"children"`
	}{Id: 0, Name: b.domain, ProcessNum: b.processNum, Children: b.taskList})
}
