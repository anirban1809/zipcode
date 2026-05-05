package view

import (
	"zipcode/src/config"

	"github.com/anirban1809/tuix/tuix"
)

func ModelSelection(props tuix.Props) tuix.Element {
	setActiveView := props.Get("setActiveView").(func(string))
	visible := props.Get("visible").(bool)

	return tuix.Box(
		tuix.Props{Direction: tuix.Column, Padding: [4]int{1, 1, 1, 1}}, tuix.NewStyle(),
		tuix.Text("Choose your model:", tuix.NewStyle()),
		tuix.Text("", tuix.NewStyle()),
		Menu(tuix.Props{Values: map[string]any{
			"items":    config.ModelNames,
			"visible":  visible,
			"viewSize": 6,
		}}, func(selected string) {
			config.CurrentModel = selected
			setActiveView("")
		}),
		tuix.Text("Press Enter to confirm, Esc to cancel", tuix.NewStyle()),
	)
}
