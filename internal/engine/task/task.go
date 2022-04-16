package task

import (
	"encoding/json"
	"fmt"
	"github.com/xiaorui77/monker-king/internal/engine/types"
	"math/rand"
	"net/url"
	"time"
)

// MainCallback 主回调函数, 仅200且无错误执行
type MainCallback func(task *Task, resp *types.ResponseWarp) error

const (
	// StateUnknown 0值
	StateUnknown       = iota // 创建时默认值
	StateScheduling           // 已经被调度
	StateRunning              // 已经开始运行
	StateInit                 // 添加之前
	StateFailed               // 运行失败
	StateSuccessful           // 运行成功
	StateSuccessfulAll        // 自己+所有子孙节点均运行成功
)

var StateStatus = map[int]string{
	0: "Unknown",
	1: "Scheduling",
	2: "Running",
	3: "Init",
	4: "Failed",
	5: "Successful",
	6: "SuccessfulAll",
}

type Task struct {
	ID       uint64                 `json:"id"`
	ParentId uint64                 `json:"pid"`
	Parent   *Task                  `json:"-"`
	Depth    int                    `json:"depth"`
	Name     string                 `json:"name"`
	State    int                    `json:"state"`
	Url      *url.URL               `json:"-"`
	Domain   string                 `json:"domain"`
	Meta     map[string]interface{} `json:"meta"`

	// 优先级: [0, MAX_INT), 值越大优先级越高
	Priority int `json:"priority"`

	Time       time.Time   `json:"createTime"`      // 创建时间
	StartTime  time.Time   `json:"startTime"` // 运行开始时间, 重试时会重置
	EndTime    time.Time   `json:"endTime"`   // 运行结束时间(保护成功和失败), 重试时会重置
	ErrDetails []ErrDetail `json:"errDetails,-"`

	Children *TaskList `json:"children"`

	// 主回调函数, 后续考虑优化合并
	Callback MainCallback `json:"-"`
}

func NewTask(name string, parent *Task, url *url.URL, fun MainCallback) *Task {
	t := &Task{
		ID:       uint64(rand.Uint32()),
		Name:     name,
		Meta:     make(map[string]interface{}, 5),
		Url:      url,
		Time:     time.Now(),
		Callback: fun,
	}
	if parent != nil {
		t.Domain = parent.Domain
		t.ParentId = parent.ID
		t.Parent = parent
		t.Depth = parent.Depth + 1
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
	switch t.State {
	case StateSuccessful:
		if t.Children != nil && t.Children.isSuccessfulAll() {
			t.State = StateSuccessfulAll
		}
	}
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

func (t *Task) IsSuccessful() bool {
	if t.State == StateSuccessfulAll || (t.Children == nil && t.State == StateSuccessful) {
		return true
	}
	return false
}

func (t *Task) refreshStatus() {
	if t.Children == nil {
		return
	}

	if t.Children.isSuccessfulAll() {
		t.State = StateSuccessfulAll
	}
}

// Push 添加子任务
// can be called by schedule.Browser
func (t *Task) Push(n *Task) {
	if t.State == StateSuccessfulAll {
		t.State = StateSuccessful
	}
	if t.Children == nil {
		t.Children = NewTaskList()
	}
	t.Children.Push(n)
}

// 获取下一个子任务
func (t *Task) nextSub() *Task {
	if t.State == StateSuccessful && t.Children != nil {
		return t.Children.Next()
	}
	return nil
}

func (t *Task) ListAll() []*Task {
	res := make([]*Task, 0, len(t.Children.list))
	res = append(res, t)

	for i := 0; i < len(res); i++ {
		task := res[i]
		if task.Children != nil {
			res = append(res, task.Children.list...)
		}
	}
	return res
}

func (t *Task) MarshalJSON() ([]byte, error) {
	type Alias Task // 否则会导致内存溢出
	return json.Marshal(&struct {
		*Alias
		Status string `json:"status"`
		Url    string `json:"url"`
	}{(*Alias)(t), t.GetState(), t.Url.String()})
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
