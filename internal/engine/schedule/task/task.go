package task

import (
	"encoding/json"
	"fmt"
	timeutil "github.com/xiaorui77/goutils/time"
	"github.com/xiaorui77/monker-king/internal/engine/types"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// MainCallback 主回调函数, 仅200且无错误执行
type MainCallback func(task *Task, resp *types.ResponseWarp) error

type OnCreated func(task *Task)

type Option func(task *Task)

const (
	// StateUnknown 0值
	StateUnknown         = iota // 创建时默认值
	StateScheduling             // 已经被调度
	StateRunning                // 已经开始运行
	StateInit                   // 添加之前
	StateFailed                 // 运行失败
	StateSuccessful             // 运行成功
	StateSuccessfulNoall        // 不完全成功
	StateCompleteNoall          // 运行完成但是没有全成功(有些重试也解决不了)
	StateSuccessfulAll          // 自己+所有子孙节点均运行成功
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
	ID       uint64 `json:"id" gorm:"primaryKey"`
	ParentId uint64 `json:"pid"`
	Parent   *Task  `json:"-" gorm:"-"`
	Depth    int    `json:"depth"`
	Name     string `json:"name"`
	State    int    `json:"state"`
	Url      string `json:"url"`
	Domain   string `json:"domain"`
	Meta     Meta   `json:"meta" gorm:"type:string"`
	// 优先级: [0, MAX_INT), 值越大优先级越高
	Priority int `json:"priority"`

	CreateTime time.Time   `json:"createTime"`          // 创建时间
	StartTime  time.Time   `json:"startTime,omitempty"` // 运行开始时间, 重试时会重置
	EndTime    time.Time   `json:"endTime,omitempty"`   // 运行结束时间(保护成功和失败), 重试时会重置
	ErrDetails []ErrDetail `json:"errDetails,omitempty"`

	Children *List `json:"children" gorm:"embedded"`

	// 主回调函数, 后续考虑优化合并
	Callback        MainCallback `json:"-" gorm:"-"`
	OnCreateHandler OnCreated    `json:"-" gorm:"-"`
}

func NewTask(name string, parent *Task, url string, fun MainCallback, opts ...Option) *Task {
	t := &Task{
		ID:         uint64(rand.Uint32()),
		Name:       name,
		Meta:       make(map[string]interface{}, 5),
		Url:        url,
		CreateTime: time.Now(),
		Callback:   fun,
	}
	if parent != nil {
		t.ParentId = parent.ID
		t.Parent = parent
		t.Domain = parent.Domain
		t.Depth = parent.Depth + 1
	}
	for _, opt := range opts {
		opt(t)
	}
	t.HandleOnCreated()
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

// SetMeta 保存K-V到meta中, 但不会立即持久化到db
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

func (t *Task) SetState(state int) {
	t.State = state
	switch t.State {
	case StateRunning:
		t.StartTime = time.Now()
		t.EndTime = timeutil.Zero
	case StateSuccessful:
		if t.State >= StateSuccessful {
			t.EndTime = time.Now()
		}
		p := t.Parent
		if p != nil && p.Children != nil && p.Children.isSuccessfulAll() {
			p.State = StateSuccessfulAll
		}
	case StateFailed:
		t.EndTime = time.Now()
	case StateSuccessfulAll:
		// 执行回调
	}
}

// RecordErr record err detail, be called after SetState(StateFailed)
func (t *Task) RecordErr(code int, msg string) {
	detail := ErrDetail{
		StartTime: t.StartTime,
		EndTime:   t.EndTime,
		Cost:      Cost(t.EndTime.Sub(t.StartTime)),
		ErrCode:   code,
		ErrMsg:    msg,
	}
	t.ErrDetails = append(t.ErrDetails, detail)
}

func (t *Task) String() string {
	return fmt.Sprintf("[%x]%s: %s", t.ID, t.Name, t.Url)
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

func (t *Task) HandleOnCreated() {
	if t.OnCreateHandler != nil {
		t.OnCreateHandler(t)
	}
}

func AddOnCreatedHandler(handler OnCreated) Option {
	return func(task *Task) {
		task.OnCreateHandler = handler
	}
}

func (t *Task) ListAll() []*Task {
	res := make([]*Task, 0, len(t.Children.Tasks))
	res = append(res, t)

	for i := 0; i < len(res); i++ {
		task := res[i]
		if task.Children != nil {
			res = append(res, task.Children.Tasks...)
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
	}{(*Alias)(t), t.GetState(), t.Url})
}
