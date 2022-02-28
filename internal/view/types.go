package view

import "github.com/rivo/tview"

type TableViewer interface {
	tview.Primitive
	TableData
}

type TableData interface {
	GetColumns() int
	GetRows() int

	GetRow(row int) []string
}

type Primitive interface {
	tview.Primitive

	// Name returns the view name.
	Name() string
}
