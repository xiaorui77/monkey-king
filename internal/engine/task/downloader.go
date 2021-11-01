package task

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/yougtao/monker-king/internal/utils"
	"io/ioutil"
	"net/http"
	"os"
)

// 下载类型任务
type downloader struct {
	fileName string
	filePath string
	url      string
}

func (t *downloader) Run(ctx context.Context, client *http.Client) error {
	logrus.Debugf("[task] downloader Run: %+v", t)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.url, nil)
	if err != nil {
		logrus.Warnf("[task] new request failed: %v", err)
		return fmt.Errorf("new request failed: %v", t.url)
	}
	req.Header = http.Header{utils.UserAgentKey: []string{utils.RandomUserAgent()}}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Warnf("[task] do request failed: %v", err)
		return fmt.Errorf("[task] do request failed: [%v]%v", http.MethodGet, t.url)
	}
	if resp.StatusCode != http.StatusOK {
		logrus.Warnf("[task] 下载失败: resp.statusCode=%v", resp.StatusCode)
		return fmt.Errorf("下载失败: 状态码=%v", resp.StatusCode)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Warnf("[task] read resp.Body failed: %v", err)
		return fmt.Errorf("read resp.Body failed")
	}

	if err := save(bytes, t.filePath, t.fileName); err != nil {
		logrus.Warnf("[task] 文件保存失败: %v", err)
		return fmt.Errorf("文件保存失败")
	}
	return nil
}

func NewDownloaderTask(name, path, url string) Task {
	return &downloader{
		fileName: name,
		filePath: path,
		url:      url,
	}
}

func save(bytes []byte, path, name string) error {
	if _, err := os.Stat(path); err != nil {
		logrus.Debugf("create path: %v", path)
		if err := os.MkdirAll(path, 0711); err != nil {
			return fmt.Errorf("create path %v failed", path)
		}
	}

	switch http.DetectContentType(bytes) {
	case "image/png":
		name += ".png"
	case "image/jpeg", "image/jpg":
		name += ".jpg"
	case "application/octet-stream":
		name += ".unknown"
	}

	filePath := fmt.Sprintf("%v/%v", path, name)
	return ioutil.WriteFile(filePath, bytes, 0666)
}
