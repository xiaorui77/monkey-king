package fileutil

import (
	"fmt"
	"github.com/xiaorui77/goutils/fileutils"
	"github.com/xiaorui77/goutils/logx"
	"io/ioutil"
	"net/http"
	"os"
)

// SaveImage 保存图片数据到指定位置
func SaveImage(bytes []byte, path, name string) error {
	name = fileutils.WindowsName(name)

	if _, err := os.Stat(path); err != nil {
		logx.Debugf("create path: %v", path)
		if err := os.MkdirAll(path, 0711); err != nil {
			return fmt.Errorf("create path %v failed: %v", path, err)
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
