package main

import (
	"log"
	"zipcode/src/ui"
	"zipcode/src/workspace"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	ui.ClearScreen()
	workspace := workspace.Load("/Users/anirban/Documents/Code/auto-code")

	p := tea.NewProgram(ui.Iniaitalize(&workspace))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	// agent debugging
	// config.HEADLESS = true
	// workspace := workspace.Load("/Users/anirban/Documents/Code/ln-api")
	// runtime := agent.NewRuntime(&workspace)
	// err := runtime.Run("list out all the dynamodb tables used in this project ")
	// if err != nil {
	// 	panic(err)
	// }
}
