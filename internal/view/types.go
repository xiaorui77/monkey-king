package view

import (
	"github.com/rivo/tview"
	"github.com/xiaorui77/monker-king/internal/view/model"
)

type TableViewer interface {
	tview.Primitive
	model.Tabular
}

// Primitive is tview.Primitive wrap
type Primitive interface {
	tview.Primitive

	// Name returns the view name.
	Name() string
}

// Igniter represents a runnable view.
type Igniter interface {
	// Start starts a component.
	Start()

	// Stop terminates a component.
	Stop()
}

type Component interface {
	Primitive
	Igniter
}
