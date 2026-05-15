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

	if skill, _ := props.Get("activeSkill").(string); skill != "" {
		status = fmt.Sprintf("Running /%s", skill)
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
		branchStyle = branchStyle.Foreground(
			tuix.Hex("#0097d7"),
		) // Yellow/orange for uncommitted changes
	} else {
		branchStyle = branchStyle.Foreground(tuix.Hex("#00c732"))
	}

	var providerName string
	var modelName string

	if config.Cfg.ActiveProviderName == "" {
		providerName = "Unconfigured"
		modelName = "Unconfigured"
	} else {
		providerName = string(config.Cfg.ActiveProviderName)
		modelName = config.Cfg.CurrentModel
	}

	line1 := tuix.Box(
		tuix.Props{Direction: tuix.Row, Justify: tuix.JustifySpaceBetween},
		tuix.NewStyle(),
		tuix.Text(
			fmt.Sprintf("%s | %s (%s)", status, workspacePath, branch),
			branchStyle,
		),
		tuix.Text(
			fmt.Sprintf(
				"Tokens:  %d\u2191 / %d\u2193 (%d)",
				inputTokens,
				outputTokens,
				totalTokens,
			),
			tuix.NewStyle(),
		),
	)

	line2 := tuix.Box(
		tuix.Props{Direction: tuix.Row, Justify: tuix.JustifySpaceBetween},
		tuix.NewStyle(),
		tuix.Text(
			fmt.Sprintf(
				"Provider: %s . Model: %s",
				providerName,
				modelName,
			),
			tuix.NewStyle(),
		),
		tuix.Text(
			fmt.Sprintf("Context: %0.2f%%", float32((totalTokens*100)/200000)),
			tuix.NewStyle(),
		),
	)

	return tuix.Box(
		tuix.Props{
			Direction: tuix.Column,
			Padding:   [4]int{1, 1, 1, 1},
			Justify:   tuix.JustifySpaceBetween,
		},
		tuix.NewStyle().Foreground(tuix.Hex("#a2a2a2")),
		tuix.Box(
			tuix.Props{Direction: tuix.Column},
			tuix.NewStyle(),
			line1, line2,
		),
	)
}
