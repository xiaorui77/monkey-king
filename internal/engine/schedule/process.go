package schedule

import (
	"context"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/monker-king/internal/engine/schedule/task"
	"time"
)

// Process 一个工作线程
type Process struct {
	browser  *Browser
	index    int                // 计数
	cancelFn context.CancelFunc // 停止函数
}

func (p *Process) run(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			logx.Errorf("[scheduler] Browser[%s] process[%d] has panic and recover: %v", p.browser.domain, p.index, err)
		}
		p.cancelFn()
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
func (p *Process) process(ctx context.Context, index int) {
	t := p.browser.next()
	if t == nil {
		logx.Debugf("[process-%d] no found tasks", index)
		return
	}
	logx.Infof("[process-%d] Task[%x] begin run, url: %s", index, t.ID, t.Url)
	p.browser.recordStart(t)

	// 设置超时并使用GET进行请求
	tCtx, cancelFunc := context.WithTimeout(ctx, p.browser.timeout(t))
	defer cancelFunc()
	resp, err := p.browser.scheduler.download.Get(tCtx, t)
	if err != nil {
		logx.Errorf("[process-%d] Task[%x] run fail, request(GET) fail: %v", index, t.ID, err)
		p.browser.recordErr(t, err.ErrCode(), err.Error())
		return
	}

	cost := time.Now().Sub(t.StartTime).Truncate(time.Millisecond * 100).Seconds()
	logx.Infof("[process-%d] Task[%x] request finish, cost: %0.1fs, will handle Callbacks", index, t.ID, cost)
	if err := p.browser.scheduler.parsing.HandleOnResponse(resp); err != nil {
		logx.Errorf("[process-%d] Task[%x] run fail, handle ResponseCallback failed: %v", index, t.ID, err)
		p.browser.recordErr(t, err.ErrCode(), err.Error())
		return
	}
	if err := t.Callback(t, resp); err != nil {
		logx.Errorf("[process-%d] Task[%x] handle task.Callback failed: %v", index, t.ID, err)
		p.browser.recordErr(t, task.ErrCallbackTask, err.Error())
		return
	}

	p.browser.recordSuccess(t)
	totalCost := t.EndTime.Sub(t.StartTime).Truncate(time.Millisecond * 100).Seconds()
	logx.Infof("[process-%d] Task[%x] run success, total cost: %0.1fs", index, t.ID, totalCost)
}
