package task

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/xiaorui77/goutils/timeutils"
	"time"
)

const (
	MetaTimeout  = "timeout" // unit: second
	MetaReader   = "reader"  // record VisualReader
	MetaSaveName = "save_name"
	MetaSavePath = "save_path"
)

type Meta map[string]interface{}

// Value implement driver.Valuer for gorm.
func (m Meta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

const (
	// ErrUnknown 0值
	ErrUnknown      = iota
	ErrNewRequest   = 512
	ErrDoRequest    = 512 + 4
	ErrReadRespBody = 1024
	ErrCallback     = 1024 + 16
	ErrCallbackTask = 1024 + 16 + 4
	ErrHttpUnknown  = 10000 // 包装http错误码
	ErrHttpNotFount = 10404 // 404页面
)

type Cost time.Duration

func (c Cost) Value() (driver.Value, error) {
	return c.Seconds(), nil
}

// Seconds 返回秒, 精确1位小数
func (c Cost) Seconds() float64 {
	return time.Duration(c).Truncate(time.Millisecond * 100).Seconds()
}

type ErrDetail struct {
	Id        uint64 `gorm:"primaryKey"`
	TaskId    uint64
	StartTime time.Time
	EndTime   time.Time
	Cost      Cost   `gorm:"type:float;precision:1"`
	ErrCode   int    `gorm:"column:code"`
	ErrMsg    string `gorm:"column:msg"`
}

func (e *ErrDetail) String() string {
	return fmt.Sprintf("ERR[%d] start:%s cost: %0.1fs msg: %s",
		e.ErrCode, e.StartTime.Format(timeutils.StampMilli), e.Cost.Seconds(), e.ErrMsg)
}
