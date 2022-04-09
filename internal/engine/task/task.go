package task

import (
	"fmt"
	"github.com/xiaorui77/monker-king/internal/engine/types"
	"github.com/xiaorui77/monker-king/internal/utils/domain"
	"math/rand"
	"net/url"
	"time"
)

// MainCallback 主回调函数, 仅200且无错误执行
type MainCallback func(task *Task, resp *types.ResponseWarp) error

const (
	StateUnknown = iota
	StateScheduling
	StateRunning
	StateInit
	StateFailed
	StateSuccessful
	StateSuccessfulAll
)

var StateStatus = map[int]string{
	0: "Unknown",
	1: "Running",
	2: "Init",
	3: "Failed",
	4: "Successful",
	5: "SuccessfulAll",
}

type Task struct {
	ID       uint64                 `json:"id"`
	ParentId uint64                 `json:"parentId"`
	Parent   *Task                  `json:"-"`
	Depth    int                    `json:"depth"`
	Name     string                 `json:"name"`
	State    int                    `json:"state"`
	Url      *url.URL               `json:"url"`
	Domain   string                 `json:"domain"`
	Meta     map[string]interface{} `json:"meta"`

	// 优先级: [0, MAX_INT), 值越大优先级越高
	Priority int `json:"priority"`

	Time       time.Time   `json:"time"`      // 创建时间
	StartTime  time.Time   `json:"startTime"` // 运行开始时间, 重试时会重置
	EndTime    time.Time   `json:"endTime"`   // 运行结束时间(保护成功和失败), 重试时会重置
	ErrDetails []ErrDetail `json:"err_details"`

	// 主回调函数, 后续考虑优化合并
	Callback MainCallback `json:"-"`
}

func NewTask(name string, parent *Task, u *url.URL, fun MainCallback) *Task {
	t := &Task{
		ID:       rand.Uint64(),
		Name:     name,
		Meta:     make(map[string]interface{}, 5),
		Url:      u,
		State:    StateUnknown,
		Time:     time.Now(),
		Callback: fun,
	}
	if parent != nil {
		t.Domain = parent.Domain
		t.ParentId = parent.ID
		t.Parent = parent
		t.Depth = parent.Depth + 1
	} else {
		t.Domain = domain.CalDomain(u)
	}
	return t
}

func (t *Task) SetPriority(p int) *Task {
	t.Priority = p
	return t
}

func (t *Task) ResetDepth() *Task {
	t.Depth = 0
	return t
}

func (t *Task) SetState(state int) {
	t.State = state
}

func (t *Task) SetMeta(key string, value interface{}) *Task {
	if key != "" && value != nil {
		t.Meta[key] = value
	}
	return t
}

func (t *Task) GetState() string {
	if s, ok := StateStatus[t.State]; ok {
		return s
	}
	return "unknown"
}

func (t *Task) RecordStart() {
	t.State = StateRunning
	t.StartTime = time.Now()
}

func (t *Task) RecordSuccess() {
	t.State = StateSuccessful
	t.EndTime = time.Now()
}

func (t *Task) RecordErr(code int, msg string) {
	t.State = StateFailed
	t.EndTime = time.Now()
	t.ErrDetails = append(t.ErrDetails, ErrDetail{
		Start:   t.StartTime,
		End:     t.EndTime,
		Cost:    t.EndTime.Sub(t.StartTime),
		ErrCode: code,
		ErrMsg:  msg,
	})
}

func (t *Task) String() string {
	return fmt.Sprintf("[%x]%s: %s", t.ID, t.Name, t.Url.String())
}

type ErrDetail struct {
	Start   time.Time
	End     time.Time
	Cost    time.Duration
	ErrCode int
	ErrMsg  string
}

func (e *ErrDetail) String() string {
	return fmt.Sprintf("ERR[%d] start:%s cost: %0.1fs msg: %s", e.ErrCode, e.Start.Format("15:04:05.000"), e.Cost.Seconds(), e.ErrMsg)
}

const (
	ErrUnknown      = iota
	ErrNewRequest   = 512
	ErrDoRequest    = 512 + 4
	ErrReadRespBody = 1024
	ErrCallback     = 1024 + 16
	ErrCallbackTask = 1024 + 16 + 4
	ErrHttpUnknown  = 10000 // 包装http错误码
	ErrHttpNotFount = 10404 // 404页面
)
