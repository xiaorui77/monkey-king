package view

import "github.com/gdamore/tcell/v2"

type Styles struct {
	Frame
	Table
	Prompt
}

type (
	Frame struct {
		BorderColor Color
	}
	Table struct {
		FgColor       Color       `yaml:"fgColor"`
		BgColor       Color       `yaml:"bgColor"`
		CursorFgColor Color       `yaml:"cursorFgColor"`
		CursorBgColor Color       `yaml:"cursorBgColor"`
		MarkColor     Color       `yaml:"markColor"`
		Header        TableHeader `yaml:"header"`
	}

	TableHeader struct {
		FgColor     Color `yaml:"fgColor"`
		BgColor     Color `yaml:"bgColor"`
		SorterColor Color `yaml:"sorterColor"`
	}

	Prompt struct {
		FgColor      Color `yaml:"fgColor"`
		BgColor      Color `yaml:"bgColor"`
		SuggestColor Color `yaml:"suggestColor"`
	}
)

type Color string

// Color returns a view color.
func (c Color) Color() tcell.Color {
	if c == "default" {
		return tcell.ColorDefault
	}

	return tcell.GetColor(string(c)).TrueColor()
}

var PresetStyles = Styles{
	Frame: Frame{
		BorderColor: "blue",
	},
	Table: Table{
		FgColor:       "#ffffff",
		BgColor:       "default",
		CursorFgColor: "#ffffff",
		CursorBgColor: "#333333",
		MarkColor:     "#f72972", // magenta
		Header: TableHeader{
			FgColor:     "#ffffff",
			BgColor:     "#333333",
			SorterColor: "blue",
		},
	},
}
