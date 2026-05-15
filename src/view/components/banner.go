package view

import "github.com/anirban1809/tuix/tuix"

func Banner(props tuix.Props) tuix.Element {
	return tuix.Box(
		tuix.Props{
			Direction: tuix.Column,
		},
		tuix.NewStyle().Foreground(tuix.Hex("#a2a2a2")),
		tuix.Text("Flux 0.0.2", tuix.NewStyle()),
		tuix.Text("Press / for options", tuix.NewStyle()),
	)
}
