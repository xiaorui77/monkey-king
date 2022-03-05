package view

import "sync"

const (
	StackPush = 1 << iota
	StackPop
)

type StackListener interface {
	StackPushed(Component)
	StackPopped(Component, Component)
}

// Stack represents a stacks of components.
type Stack struct {
	components []Component
	listeners  []StackListener
	mx         sync.RWMutex
}

func NewStack() *Stack {
	return &Stack{
		components: make([]Component, 0, 2),
		listeners:  make([]StackListener, 0, 2),
	}
}

func (s *Stack) AddListener(l StackListener) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.listeners = append(s.listeners, l)
}

// Top returns the top most item
func (s *Stack) Top() Component {
	if len(s.components) == 0 {
		return nil
	}
	return s.components[len(s.components)-1]
}

func (s *Stack) Push(c Component) {
	if top := s.Top(); top != nil {
		top.Stop()
	}

	s.mx.Lock()
	s.components = append(s.components, c)
	s.mx.Unlock()
	s.notify(c, StackPush)
}

func (s *Stack) Pop() Component {
	if len(s.components) == 0 {
		return nil
	}

	s.mx.Lock()
	c := s.components[len(s.components)-1]
	s.components = s.components[:len(s.components)-1]
	s.mx.Unlock()

	s.notify(c, StackPop)
	return c
}

func (s *Stack) Clear() {
	for range s.components {
		s.Pop()
	}
}

func (s *Stack) notify(c Component, action int) {
	for _, l := range s.listeners {
		switch action {
		case StackPush:
			l.StackPushed(c)
		case StackPop:
			l.StackPopped(c, s.Top())
		}
	}
}
