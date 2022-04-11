package task

import (
	"net/url"
	"testing"
)

var root *Task

func setup() {
	u, _ := url.Parse("https://example.com")
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
	t.Logf("%v", root.children.ListAll())
}
