package task

import (
	"encoding/json"
	"github.com/xiaorui77/goutils/logx"
	"sync"
)

type Node struct {
	Task *Task

	Parent   *Node
	Children *List
}

func NewNode(task *Task) *Node {
	return &Node{
		Task: task,
	}
}

func (n *Node) push(node *Node) {
	if n.Children == nil {
		n.Children = &List{nodes: make([]*Node, 0, 10)}
	}
	n.Children.push(node)
	if n.Task.State == StateSuccessfulAll {
		n.Task.SetState(StateSuccessful)
	}
}

func (n *Node) next() *Node {
	if n.Children != nil {
		return n.Children.next()
	}
	return nil
}

func (n *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id       uint64 `json:"id"`
		ParentId uint64 `json:"pid"`
		Name     string `json:"name"`
		Status   string `json:"status"`
		Children *List  `json:"children"`
	}{Id: n.Task.ID, ParentId: n.Task.ParentId, Name: n.Task.Name, Status: n.Task.GetState(), Children: n.Children})
}

type List struct {
	mu     sync.Mutex
	nodes  []*Node
	offset int
}

func (l *List) push(node *Node) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := len(l.nodes) - 1; i >= 0; i-- {
		if node.Task.Priority <= l.nodes[i].Task.Priority {
			l.nodes = append(l.nodes, nil)
			copy((l.nodes)[i+2:], (l.nodes)[i+1:])
			l.nodes[i+1] = node

			if l.offset > i+1 {
				l.offset = i + 1
			}
			return
		}
	}
	// 插入前部
	l.nodes = append([]*Node{node}, l.nodes...)
	l.offset = 0
}

func (l *List) next() *Node {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i := 0; i < len(l.nodes); i++ {
		j := (l.offset + i) % len(l.nodes)
		if l.nodes[j].Task.State == StateInit {
			l.offset = j + 1
			l.nodes[j].Task.SetState(StateScheduling)
			return l.nodes[j]
		}
	}
	for i := 0; i < len(l.nodes); i++ {
		if n := l.nodes[i].next(); n != nil {
			return n
		}
	}
	return nil
}

func (l *List) delete(node *Node) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i := range l.nodes {
		if l.nodes[i] == node {
			l.nodes = append(l.nodes[:i], l.nodes[i+1:]...)
			return
		}
	}
}

func (l *List) MarshalJSON() ([]byte, error) {
	if l != nil {
		return json.Marshal(l.nodes)
	}
	return json.Marshal([]struct{}{})
}

type Tree struct {
	Domain  string
	root    *List
	queue   []*Node // 无须保存, 方便查询
	queueMu sync.RWMutex
}

func NewTree(domain string) *Tree {
	return &Tree{
		Domain: domain,
		root: &List{
			nodes: make([]*Node, 0, 10),
		},
		queue: make([]*Node, 0, 10),
	}
}

// Next 返回下个可运行的任务, 默认使用深度优先遍历
func (tree *Tree) Next() *Task {
	if n := tree.root.next(); n != nil {
		return n.Task
	}
	return nil
}

func (tree *Tree) Push(task *Task) {
	if task == nil {
		return
	}
	node := NewNode(task)
	tree.queue = append(tree.queue, node)
	if parent := tree.findNode(task.Parent); parent != nil {
		node.Parent = parent
		parent.push(node)
	} else {
		tree.root.push(node)
	}
}

// Refresh 分析fail状态的task, 转为init
func (tree *Tree) Refresh() {
	tree.queueMu.Lock()
	defer tree.queueMu.Unlock()

	for _, node := range tree.queue {
		t := node.Task
		if t.State == StateFailed && len(t.ErrDetails) > 0 {
			if len(t.ErrDetails) > 7 {
				continue // 超过7次不再重试
			}
			n := 0
			for i := len(t.ErrDetails) - 1; i >= 0; i-- {
				if t.ErrDetails[i].ErrCode == ErrHttpNotFount {
					n++
				} else {
					break
				}
			}
			if n < 2 {
				// 连续的NotFound错误小于2次才重试
				logx.Infof("[browser] Task[%x] can be retry, last err: %s", t.ID, t.ErrDetails[len(t.ErrDetails)-1])
				t.SetState(StateInit)
				if node.Parent != nil {
					node.Parent.Children.offset = 0
				}

			}
		}
	}
}

func (tree *Tree) findNode(task *Task) *Node {
	if task == nil {
		return nil
	}
	tree.queueMu.RLock()
	defer tree.queueMu.RUnlock()
	for i := range tree.queue {
		if tree.queue[i].Task.ID == task.ID {
			return tree.queue[i]
		}
	}
	return nil
}

// must mutex by caller
func (tree *Tree) find(id uint64) (*Node, int) {
	for i := range tree.queue {
		if tree.queue[i].Task.ID == id {
			return tree.queue[i], i
		}
	}
	return nil, -1
}

func (tree *Tree) MarshalJSON() ([]byte, error) {
	if tree.root != nil {
		res := struct {
			Id       uint64 `json:"id"`
			Name     string `json:"name"`
			Children *List  `json:"children"`
		}{Id: 0, Name: tree.Domain, Children: tree.root}
		return json.Marshal(res)
	}
	return json.Marshal(nil)
}

func (tree *Tree) Find(id uint64) *Task {
	tree.queueMu.RLock()
	defer tree.queueMu.RUnlock()
	node, _ := tree.find(id)
	return node.Task
}

func (tree *Tree) Query(name string) *Task {
	tree.queueMu.RLock()
	defer tree.queueMu.RUnlock()
	for i := range tree.queue {
		if tree.queue[i].Task.Name == name {
			return tree.queue[i].Task
		}
	}
	return nil
}

// Delete 删除一个节点
// todo: 只能删除叶子节点
func (tree *Tree) Delete(id uint64) *Task {
	tree.queueMu.Lock()
	defer tree.queueMu.Unlock()

	node, i := tree.find(id)
	if node != nil {
		if node.Parent != nil {
			node.Parent.Children.delete(node)
		} else {
			tree.root.delete(node)
		}
		tree.queue = append(tree.queue[:i], tree.queue[i+1:]...)
		return node.Task
	}
	return nil
}

func (tree *Tree) List() []*Task {
	tree.queueMu.RLock()
	defer tree.queueMu.RUnlock()
	res := make([]*Task, 0, len(tree.queue))
	for i := range tree.queue {
		res = append(res, tree.queue[i].Task)
	}
	return res
}

func (tree *Tree) GetTree() *Tree {
	return tree
}
