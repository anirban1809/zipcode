package view

import (
	view "zipcode/src/ui/components/utils"

	"github.com/anirban1809/tuix/tuix"
)

func Skills(props tuix.Props) tuix.Element {
	return tuix.Box(tuix.Props{Direction: tuix.Column, Padding: [4]int{1, 1, 1, 1}},
		tuix.NewStyle(),
		tuix.Text("All Skills", tuix.NewStyle()),
		tuix.Text("No skills found", tuix.NewStyle()),
		view.NewLine(),
		tuix.Text("Create skills in .flux/skills/ or ~/.flux/skills/", tuix.NewStyle()),
	)
}
