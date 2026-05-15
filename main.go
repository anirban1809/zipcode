package main

import (
	"log"
	"os"
	"zipcode/src/agent"
	"zipcode/src/config"
	"zipcode/src/utils"
	"zipcode/src/view"
	"zipcode/src/workspace"

	"github.com/anirban1809/tuix/tuix"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}


	//------------bubbletea implementation (deprecated)----------//
	// dir, err := os.Getwd()

	// if err != nil {
	// 	log.Fatal("Failed to get current directory. Error: ", err.Error())
	// }

	// workspace := workspace.Load(dir)

	// p := tea.NewProgram(ui.Iniaitalize(&workspace), tea.WithAltScreen())
	// if _, err := p.Run(); err != nil {
	// 	log.Fatal(err)
	// }

	//---------------tuix implementation-----------------//
	width, height, _ := utils.GetTerminalSize()
	dir, _ := os.Getwd()
	ws := workspace.Load(dir)

	runtime := agent.NewRuntime(&ws)

	app := tuix.NewApp(width, height)
	app.Run(view.App, tuix.Props{Values: map[string]any{"runtime": &runtime, "wd": dir}})

	// agent debugging
	// config.Cfg.Headless = true
	// // dir, err := os.Getwd()

	// // if err != nil {
	// // 	log.Fatal("Failed to get current directory. Error: ", err.Error())
	// // }

	// workspace := workspace.Load("/Users/anirban/Documents/Code/zipcode")
	// runtime := agent.NewRuntime(&workspace)
	// _, err := runtime.Run("Read and analyze app.go file")
	// // _, err := runtime.Run("The image routes randomly fail with a 500 error. Diagnose the issue and fix it.")
	// // runtime.Run("What version of typescript is being used in this project?")
	// // runtime.Run("Add the above information in a file packages.md")
	// if err != nil {
	// 	panic(err)
	// }
}
