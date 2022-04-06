package download

import (
	"context"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/monker-king/internal/engine/task"
	"github.com/xiaorui77/monker-king/internal/utils"
	"github.com/xiaorui77/monker-king/pkg/error"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const (
	MaxTimeout = time.Minute * 10
)

type Downloader struct {
	client *http.Client
}

func NewDownloader(_ context.Context) *Downloader {
	jar, err := cookiejar.New(nil)
	if err != nil {
		logx.Errorf("[downloader] new cookiejar failed: %v", err)
		return nil
	}

	return &Downloader{
		client: &http.Client{
			Jar: jar,
			// The timeout includes connection time, any redirects, and reading the response body.
			// includes Dial、TLS handshake、Request、Resp.Headers、Resp.Body, excludes Idle
			Timeout: MaxTimeout,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   15 * time.Second,
					KeepAlive: 10 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:   true,
				TLSHandshakeTimeout: 15 * time.Second,
				IdleConnTimeout:     60 * time.Second,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
			},
		},
	}
}

// Get send an HTTP Request by GET Method.
// Caller should close resp.Body when done reading from it.
func (d *Downloader) Get(ctx context.Context, t *task.Task) (*http.Request, *http.Response, error.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.Url.String(), nil)
	if err != nil {
		logx.Errorf("[downloader] request.Get failed: %v")
		return nil, nil, &error.Err{Err: err, Code: task.ErrNewRequest}
	}
	d.beforeReq(req)

	resp, err := d.client.Do(req)
	if err != nil {
		logx.Warnf("[downloader] request.Do failed: %v", err)
		return nil, nil, &error.Err{Code: task.ErrDoRequest, Err: err}
	}
	return req, resp, nil
}

func (d *Downloader) beforeReq(req *http.Request) {
	req.Header.Set(utils.UserAgentKey, utils.RandomUserAgent())

	// TODO: 待设定
	// req.Close = true
}
