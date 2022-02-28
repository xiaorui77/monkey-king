package view

import "sync"

// Stack represents a stacks of components.
type Stack struct {
	components []Primitive
	listeners  []StackListener
	mx         sync.RWMutex
}

type StackListener interface {
	StackPushed()
	StackPopped()
}
