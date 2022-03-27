package download

import (
	"context"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/monker-king/internal/engine/task"
	"github.com/xiaorui77/monker-king/internal/utils"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type Downloader struct {
	client *http.Client
	ctx    context.Context
}

func NewDownloader(ctx context.Context) *Downloader {
	jar, err := cookiejar.New(nil)
	if err != nil {
		logx.Errorf("[scheduler] new cookiejar failed: %v", err)
		return nil
	}

	return &Downloader{
		ctx: ctx,
		client: &http.Client{
			Jar:     jar,
			Timeout: time.Second * 15,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   15 * time.Second,
					KeepAlive: 10 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   10,
				IdleConnTimeout:       60 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

func (d *Downloader) Get(t *task.Task) {
	req, err := http.NewRequestWithContext(d.ctx, http.MethodGet, t.Url.String(), nil)
	if err != nil {
		logx.Errorf("[download] request.Get failed: %v")
		t.SetState(task.StateFail)
		return
	}
	d.beforeReq(req)

	resp, err := d.client.Do(req)
	if err != nil {
		logx.Warnf("[download] request.Do failed: %v")
		t.HandleOnResponseErr(resp, err)
		return
	}

	t.HandleOnResponse(req, resp)
}

func (d *Downloader) beforeReq(req *http.Request) {
	req.Header.Set(utils.UserAgentKey, utils.RandomUserAgent())

	// TODO: 待设定
	// req.Close = true
}
