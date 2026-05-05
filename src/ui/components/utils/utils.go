package view

import "github.com/anirban1809/tuix/tuix"

// must be used in a box with column direction
func NewLine() tuix.Element {
	return tuix.Text("", tuix.NewStyle())
}
