package download

import (
	"context"
	"fmt"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/monker-king/internal/engine/schedule/task"
	"github.com/xiaorui77/monker-king/internal/engine/types"
	"github.com/xiaorui77/monker-king/internal/utils"
	"github.com/xiaorui77/monker-king/internal/utils/fileutil"
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

func NewDownloader() *Downloader {
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
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   10,
				IdleConnTimeout:       60 * time.Second,
				TLSHandshakeTimeout:   15 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

// Get send an HTTP Request by GET Method.
// Caller should close resp.Body when done reading from it.
func (d *Downloader) Get(ctx context.Context, t *task.Task) (*types.ResponseWarp, error.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.Url, nil)
	if err != nil {
		logx.Errorf("[downloader] Task[%08x] new request failed: %v", t.ID, err)
		return nil, &error.Err{Err: err, Code: task.ErrNewRequest}
	}
	reqWrap := &types.RequestWrap{
		URL:     req.URL,
		BaseURL: req.URL,
	}
	d.beforeReq(req)

	logx.Debugf("[downloader] Task[%08x] send request, header: %v", t.ID, req.Header)
	resp, err := d.client.Do(req)
	if err != nil {
		logx.Warnf("[downloader] Task[%08x] do request failed: %v", t.ID, err)
		return nil, &error.Err{Code: task.ErrDoRequest, Err: err}
	}
	logx.Debugf("[downloader] Task[%08x] request done, response header: %v", t.ID, resp.Header)

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logx.Errorf("[downloader] resp.Body close fail: %v", err)
		}
	}()
	reader := &fileutil.VisualReader{
		Reader: resp.Body,
		Total:  resp.ContentLength,
	}
	body, err := reader.ReadAll()
	if err != nil {
		t.SetMeta(task.MetaReader, reader) // convention：如果有错误，则记录reader
		return nil, &error.Err{
			Code: task.ErrReadRespBody,
			Err:  fmt.Errorf("reading resp.Body when[%v/%v] failed: %v", reader.Cur, reader.Total, err),
		}
	}

	return &types.ResponseWarp{
		StatusCode: resp.StatusCode,
		Body:       body,
		Request:    reqWrap,
	}, nil
}

func (d *Downloader) beforeReq(req *http.Request) {
	req.Header.Set(utils.UserAgentKey, utils.RandomUserAgent())

	req.Header.Set("accept-encoding", "")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")

	// TODO: 待设定
	// req.Close = true
}
