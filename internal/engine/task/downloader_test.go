package task

import (
	"context"
	"testing"
)

func TestDownload(t *testing.T) {
	// url := "https://img-blog.csdnimg.cn/20191220155031214.png"
	url := "https://i.hexuexiao.cn/up/e0/4e/ba/712d31981a9b9de00657f4b97eba4ee0.jpg"
	task := NewDownloaderTask("test", "./", url)
	if err := task.Run(context.TODO()); err != nil {
		t.Errorf("download failed: %v", err)
	}
}
