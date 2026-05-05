package view

import (
	"fmt"
	"zipcode/src/config"

	"github.com/anirban1809/tuix/tuix"
)

func StatusLine(props tuix.Props) tuix.Element {
	workspacePath, _ := props.Get("workspacePath").(string)
	status := "Idle"
	if running, _ := props.Get("running").(bool); running {
		status = "Running"
	}

	return tuix.Box(
		tuix.Props{Direction: tuix.Row, Padding: [4]int{0, 1, 1, 1}, Justify: tuix.JustifySpaceBetween},
		tuix.NewStyle().Foreground(tuix.Hex("#a2a2a2")),
		tuix.Text(
			fmt.Sprintf("%s | %s", status, workspacePath),
			tuix.NewStyle(),
		),
		tuix.Text(
			fmt.Sprintf("Model: %s", config.CurrentModel),
			tuix.NewStyle(),
		),
	)
}
