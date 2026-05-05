package main

import (
	"os"
	"zipcode/src/agent"
	"zipcode/src/utils"
	"zipcode/src/view"
	"zipcode/src/workspace"

	"github.com/anirban1809/tuix/tuix"
)

func main() {

	// dir, err := os.Getwd()

	// if err != nil {
	// 	log.Fatal("Failed to get current directory. Error: ", err.Error())
	// }

	// workspace := workspace.Load(dir)

	// p := tea.NewProgram(ui.Iniaitalize(&workspace), tea.WithAltScreen())
	// if _, err := p.Run(); err != nil {
	// 	log.Fatal(err)
	// }

	//--------------------------------------//
	width, height, _ := utils.GetTerminalSize()
	dir, _ := os.Getwd()
	ws := workspace.Load(dir)

	runtime := agent.NewRuntime(&ws)

	app := tuix.NewApp(width, height)
	app.Run(view.App, tuix.Props{Values: map[string]any{"runtime": &runtime, "wd": dir}})

	// agent debugging
	// config.HEADLESS = true
	// // dir, err := os.Getwd()

	// // if err != nil {
	// // 	log.Fatal("Failed to get current directory. Error: ", err.Error())
	// // }

	// workspace := workspace.Load("/Users/anirban/Documents/Code/example-api")
	// runtime := agent.NewRuntime(&workspace)
	// _, err := runtime.Run("Create a file test11.md and add some example text in it")
	// // _, err := runtime.Run("The image routes randomly fail with a 500 error. Diagnose the issue and fix it.")
	// // runtime.Run("What version of typescript is being used in this project?")
	// // runtime.Run("Add the above information in a file packages.md")
	// if err != nil {
	// 	panic(err)
	// }
}
