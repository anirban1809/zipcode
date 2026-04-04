package main

import (
	"log"
	"os"
	"zipcode/src/ui"
	"zipcode/src/workspace"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	ui.ClearScreen()
	dir, err := os.Getwd()

	if err != nil {
		log.Fatal("Failed to get current directory. Error: ", err.Error())
	}

	workspace := workspace.Load(dir)

	p := tea.NewProgram(ui.Iniaitalize(&workspace), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	// agent debugging
	// config.HEADLESS = true
	// // dir, err := os.Getwd()

	// // if err != nil {
	// // 	log.Fatal("Failed to get current directory. Error: ", err.Error())
	// // }

	// workspace := workspace.Load("/Users/anirban/Documents/Code/example-api")
	// runtime := agent.NewRuntime(&workspace)
	// err := runtime.Run("Find the index.ts file, Read it and create a summary")
	// // runtime.Run("What version of typescript is being used in this project?")
	// // runtime.Run("Add the above information in a file packages.md")
	// if err != nil {
	// 	panic(err)
	// }
}
