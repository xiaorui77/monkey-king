package task

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// downloader 是下载类型任务
type downloader struct {
	fileName string
	filePath string
	url      string
}

var DefaultClient = &http.Client{
	Timeout: time.Second * 30,
}

func (t downloader) Run(ctx context.Context) error {
	logrus.Debugf("[task] downloader Run: %+v", t)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.url, nil)
	if err != nil {
		logrus.Warnf("[task] new request failed: %v", err)
		return fmt.Errorf("new request failed: %v", t.url)
	}
	resp, err := DefaultClient.Do(req)
	if err != nil {
		logrus.Warnf("[task] do request failed: %v", err)
		return fmt.Errorf("[task] do request failed: [%v]%v", http.MethodGet, t.url)
	}
	if resp.StatusCode != http.StatusOK {
		logrus.Warnf("[task] 下载失败: resp.statusCode=%v", resp.StatusCode)
		return fmt.Errorf("下载失败: 状态码=%v", resp.StatusCode)
	}

	if err := save(resp.Body, t.filePath, t.fileName); err != nil {
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

func save(data io.Reader, path, name string) error {
	if _, err := os.Stat(path); err != nil {
		logrus.Debugf("create path: %v", path)
		if err := os.MkdirAll(path, 0711); err != nil {
			return fmt.Errorf("create path %v failed", path)
		}
	}

	bytes, err := ioutil.ReadAll(data)
	if err != nil {
		logrus.Warnf("[task] 读取图片内容时出错: %v", err)
		return fmt.Errorf("读取图片内容时出错")
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
