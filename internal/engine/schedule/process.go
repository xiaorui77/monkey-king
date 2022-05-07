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

	logger *logx.Entry
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
			p.logger.Infof("[scheduler] Browser[%s] Process[%d] has been stopped", p.browser.domain, p.index)
			return
		default:
			p.process(ctx)
		}
		time.Sleep(time.Second * TaskInterval)
	}
}

// process a task
func (p *Process) process(ctx context.Context) {
	logger := p.logger.WithField("index", p.index)
	t := p.browser.next()
	if t == nil {
		logger.Debugf("[process-%d] no found tasks", p.index)
		return
	}
	timeout := p.browser.timeout(t)
	logx.Infof("[process-%d] Task[%x] begin run, timeout: %0.1fs, url: %s", p.index, t.ID, timeout.Seconds(), t.Url)
	p.browser.recordStart(t)

	// 设置超时并使用GET进行请求
	tCtx, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()
	resp, err := p.browser.scheduler.download.Get(tCtx, t)
	if err != nil {
		cost := time.Now().Sub(t.StartTime).Truncate(time.Millisecond * 100).Seconds()
		logx.Errorf("[process-%d] Task[%x] run failed, cost: %0.1fs, request(GET) fail: %v", p.index, t.ID, cost, err)
		p.browser.recordErr(t, err.ErrCode(), err.Error())
		return
	}

	cost := time.Now().Sub(t.StartTime).Truncate(time.Millisecond * 100).Seconds()
	logx.Infof("[process-%d] Task[%x] request finish, cost: %0.1fs, will handle Callbacks", p.index, t.ID, cost)
	if err := p.browser.scheduler.parsing.HandleOnResponse(resp); err != nil {
		logx.Errorf("[process-%d] Task[%x] run failed, handle ResponseCallback failed: %v", p.index, t.ID, err)
		p.browser.recordErr(t, err.ErrCode(), err.Error())
		return
	}
	if err := t.Callback(t, resp); err != nil {
		logx.Errorf("[process-%d] Task[%x] run failed, handle task.Callback failed: %v", p.index, t.ID, err)
		p.browser.recordErr(t, task.ErrCallbackTask, err.Error())
		return
	}

	p.browser.recordSuccess(t)
	totalCost := t.EndTime.Sub(t.StartTime).Seconds()
	logx.Infof("[process-%d] Task[%x] run success, total cost: %0.1fs", p.index, t.ID, totalCost)
}
