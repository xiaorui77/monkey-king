package view

import (
	"github.com/rivo/tview"
	"github.com/yougtao/monker-king/internal/view/model"
)

type TableViewer interface {
	tview.Primitive
	model.Tabular
}

// 类比 Component
type Primitive interface {
	tview.Primitive

	// Name returns the view name.
	Name() string
}

type Component interface {
	Primitive

	Start()
}

// ------------------------------------------------------------
// table data
