package task

import (
	"testing"
)

var root *Task

func setup() {
	u := "https://example.com"
	root = NewTask("root", nil, u, nil)
	a1 := NewTask("a1", root, u, nil)
	a2 := NewTask("a2", root, u, nil)
	a3 := NewTask("a3", root, u, nil)
	b1 := NewTask("b1", a1, u, nil)
	b2 := NewTask("b2", a1, u, nil)
	root.Push(a1)
	root.Push(a2)
	root.Push(a3)
	a1.Push(b1)
	a1.Push(b2)
}

func TestListAll(t *testing.T) {
	setup()
	t.Logf("%v", root.ListAll())
	t.Logf("%v", root.Children.ListAll())
}

func TestNext(t *testing.T) {
	u := "https://example.com"
	a := NewTask("a", nil, u, nil)
	a.SetState(StateInit)
	b := NewTask("b", nil, u, nil)
	b.SetState(StateInit)

	a1 := NewTask("a1", a, u, nil)
	a1.SetState(StateInit)
	a.Push(a1)
	a2 := NewTask("a2", a, u, nil)
	a2.SetState(StateInit)
	a.Push(a2)
	a3 := NewTask("a3", a, u, nil)
	a3.SetState(StateInit)
	a.Push(a3)
	b1 := NewTask("b1", b, u, nil)
	b1.SetState(StateInit)
	b.Push(b1)
	b2 := NewTask("b2", b, u, nil)
	b2.SetState(StateInit)
	b.Push(b2)

	taskList := NewTaskList()
	taskList.Push(a)
	taskList.Push(b)

	for i := 0; ; i++ {
		ta := taskList.Next()
		if ta == nil {
			break
		}
		t.Logf("%dth: %v", i, ta)
	}
}
