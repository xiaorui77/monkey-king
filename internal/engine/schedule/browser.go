package schedule

import (
	"context"
	"github.com/rfyiamcool/backoff"
	"github.com/yougtao/goutils/logx"
	"github.com/yougtao/goutils/wait"
	"math/rand"
	"sync"
	"time"
)

// DomainBrowser 在同一个域名下的调度器
// 1. 处理同一个Domain下的优先级关系
// 2. 管理cookie等
type DomainBrowser struct {
	domain string

	priority *TaskQueue
	normal   *TaskQueue
}

func NewDomainBrowser(host string) *DomainBrowser {
	return &DomainBrowser{
		domain:   host,
		priority: NewTaskQueue(),
		normal:   NewTaskQueue(),
	}
}

func (d *DomainBrowser) Push(priority bool, task *Task) {
	if priority {
		d.priority.push(task)
	} else {
		d.normal.push(task)
	}
}

func (d *DomainBrowser) Next() *Task {
	if task := d.priority.next(); task != nil {
		return task
	}
	return d.normal.next()
}

func (d *DomainBrowser) List() []*Task {
	res := make([]*Task, 0, len(d.priority.list())+len(d.normal.list()))
	res = append(res, d.priority.list()...)
	res = append(res, d.normal.list()...)
	return res
}

// Begin begin schedule all tasks by multi-thread.
func (d *DomainBrowser) Begin(ctx context.Context) {
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
	// 退避
	sh := backoff.NewBackOff(
		backoff.WithMinDelay(2*time.Second),
		backoff.WithMaxDelay(15*time.Second),
		backoff.WithFactor(2),
	)

	for {
		select {
		case <-ctx.Done():
			logx.Infof("[schedule] The process-%d[%s] will stop", index, d.domain)
			wg.Done()
			return
		default:
			task := d.Next()
			if task == nil {
				// 退避
				sh.SleepCtx(ctx)
				continue
			}

			if task.ID == 0 {
				task.ID = rand.Uint64()
			}
			logx.Infof("[schedule] The schedule[%x] begin to run, url: %s", task.ID, task.url)
			task.state = TaskStateRunning
			if err := task.Run(ctx, httpClient); err != nil {
				task.SetState(TaskStateFail)
				logx.Warnf("[schedule] The schedule[%x] run failed(try again after 5s): %v", task.ID, err)
				// 重试
				d.Push(true, task)
				time.Sleep(time.Second * 5)
				continue
			}
			sh.Reset()
			task.SetState(TaskStateSuccess)
			logx.Infof("[schedule] The schedule[%x] done.", task.ID)
			// 成功时的延迟
			time.Sleep(time.Second * 2)
		}
	}
}

func (d *DomainBrowser) reFail(_ *Task) {

}

func (d *DomainBrowser) close() {
	// todo: close all task queue of the domain
}

func (d *DomainBrowser) MarshalJSON() ([]byte, error) {
	return nil, nil
}
