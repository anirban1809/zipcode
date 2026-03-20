package main

import (
	"log"
	"zipcode/src/ui"
	"zipcode/src/workspace"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	ui.ClearScreen()
	workspace := workspace.Load("/Users/anirban/Documents/Code/ln-api")

	p := tea.NewProgram(ui.Iniaitalize(&workspace), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	// agent debugging
	// config.HEADLESS = true
	// workspace := workspace.Load("/Users/anirban/Documents/Code/ln-api")
	// runtime := agent.NewRuntime(&workspace)
	// runtime.Run("Create a detailed description about all the API routes implemented in this project. Store this info in a file API_Routes.md")
	// // runtime.Run("What version of typescript is being used in this project?")
	// // runtime.Run("Add the above information in a file packages.md")
	// // if err != nil {
	// // 	panic(err)
	// // }
}
