package localfile

import (
	"fmt"
	"github.com/xiaorui77/goutils/logx"
	"io/ioutil"
	"net/http"
	"os"
)

// SaveImage 保存图片数据到指定位置
func SaveImage(bytes []byte, path, name string) error {
	if _, err := os.Stat(path); err != nil {
		logx.Debugf("create path: %v", path)
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
	default:
		name += ".unknown"
	}

	filePath := fmt.Sprintf("%v/%v", path, name)
	return ioutil.WriteFile(filePath, bytes, 0666)
}
