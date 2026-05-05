package view

import (
	"os"
	"strings"
	"zipcode/src/utils"
	view "zipcode/src/view/components"

	"github.com/anirban1809/tuix/tuix"
)

type CommandKind int

const (
	CmdView CommandKind = iota
	CmdPrompt
	CmdAction
)

type Command struct {
	Name   string
	Kind   CommandKind
	Prompt string
	Run    func()
}

func MainMenu(props tuix.Props) tuix.Element {
	activeView := props.Get("activeView").(string)
	setActiveView := props.Get("setActiveView").(func(string))
	submitPrompt := props.Get("submitPrompt").(func(string))
	prompt := props.Get("prompt").(string)

	var commands = []Command{
		{Name: "/models", Kind: CmdView},
		{Name: "/skills", Kind: CmdView},
		{Name: "/agents", Kind: CmdView},
		{Name: "/settings", Kind: CmdView},
		{Name: "/about", Kind: CmdPrompt, Prompt: "Tell me about this project."},
		{Name: "/exit", Kind: CmdAction, Run: func() { os.Exit(0) }},
		{Name: "/clear", Kind: CmdAction, Run: func() { /* clear outputs */ }},
	}

	findCommand := func(selected string) Command {
		var toFind Command
		for _, command := range commands {
			if command.Name == selected {
				toFind = command
				break
			}
		}
		return toFind
	}

	filteredItems := utils.Filter(commands, func(item Command, index int) bool {
		return strings.HasPrefix(item.Name, prompt)
	})

	if activeView != "" && tuix.CurrentKey.Code == tuix.KeyEscape {
		setActiveView("")
	}

	modelSelection := view.ModelSelection(tuix.Props{Values: map[string]any{
		"setActiveView": setActiveView,
		"visible":       activeView == "/models",
	}})

	skillsView := view.Skills(tuix.Props{})
	agentsView := view.Agent(tuix.Props{})

	if activeView == "/models" {
		return modelSelection
	}

	if activeView == "/exit" {
		os.Exit(0)
	}

	if activeView == "/skills" {
		return skillsView
	}

	if activeView == "/agents" {
		return agentsView
	}

	commandNames := utils.Map(filteredItems, func(item Command, index int) string {
		return item.Name
	})

	return view.Menu(tuix.Props{
		Values: map[string]any{
			"items":   commandNames,
			"visible": activeView == "",
		},
	}, func(selected string) {
		cmd := findCommand(selected) // lookup in `commands`
		switch cmd.Kind {
		case CmdView:
			setActiveView(cmd.Name)
		case CmdPrompt:
			submitPrompt(cmd.Prompt)
			setActiveView("") // dismiss menu
		case CmdAction:
			cmd.Run()
		}
	})
}
