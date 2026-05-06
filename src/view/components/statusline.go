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

	branch := "main"
	if v, ok := props.Get("branch").(string); ok && v != "" {
		branch = v
	}

	hasUncommittedChanges := props.Get("hasUncommittedChanges").(bool)

	inputTokens := 0
	outputTokens := 0
	if v, ok := props.Get("inputTokens").(int); ok {
		inputTokens = v
	}
	if v, ok := props.Get("outputTokens").(int); ok {
		outputTokens = v
	}
	totalTokens := inputTokens + outputTokens

	branchStyle := tuix.NewStyle()
	if hasUncommittedChanges {
		branchStyle = branchStyle.Foreground(tuix.Hex("#0097d7")) // Yellow/orange for uncommitted changes
	} else {
		branchStyle = branchStyle.Foreground(tuix.Hex("#00c732"))
	}

	return tuix.Box(
		tuix.Props{Direction: tuix.Row, Padding: [4]int{0, 1, 1, 1}, Justify: tuix.JustifySpaceBetween},
		tuix.NewStyle().Foreground(tuix.Hex("#a2a2a2")),
		tuix.Text(
			fmt.Sprintf("%s | %s (%s)", status, workspacePath, branch),
			branchStyle,
		),
		tuix.Box(
			tuix.Props{Direction: tuix.Row, Gap: 2},
			tuix.NewStyle(),
			tuix.Text(
				fmt.Sprintf("Model: %s", config.CurrentModel),
				tuix.NewStyle(),
			),
			tuix.Text(
				fmt.Sprintf("Tokens:  %d\u2191 / %d\u2193 (%d), Context: %0.2f%%", inputTokens, outputTokens, totalTokens, float64(totalTokens*100)/200000),
				tuix.NewStyle(),
			),
		),
	)
}
