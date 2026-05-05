package view

import (
	"fmt"
	"strings"

	"zipcode/src/agent"
	view "zipcode/src/view/components"
	"zipcode/src/workspace"

	"github.com/anirban1809/tuix/tuix"
	"github.com/anirban1809/tuix/tuix/components"
)

var promptChan = make(chan string)

func App(props tuix.Props) tuix.Element {
	prompt, setPrompt := tuix.UseState("")
	activeSession, setActiveSession := tuix.UseState(false)
	outputs, setOutputs := tuix.UseState([]tuix.Element{})
	livePrompt, setLivePrompt := tuix.UseState("")
	livePromptIdx, setLivePromptIdx := tuix.UseState(-1)
	activeMenuView, setActiveMenuView := tuix.UseState("")

	questionVisible, setQuestionVisible := tuix.UseState(false)
	question, setQuestion := tuix.UseState(struct {
		question string
		options  []string
	}{})
	selectedOption, setSelectedOption := tuix.UseState("")
	optionSelected, setOptionSelected := tuix.UseState(false)
	fileDiff, setFileDiff := tuix.UseState(agent.FileChangeEvent{})

	runtime := props.Get("runtime").(*agent.Runtime)

	menuVisible := strings.HasPrefix(prompt, "/")

	submitPrompt := func(p string) {
		promptChan <- p
		setPrompt("")
		if !activeSession {
			setActiveSession(true)
		}
		go runtime.Run(p)
	}

	if tuix.CurrentKey.Code == tuix.KeyEnter && !activeSession && !menuVisible {
		submitPrompt(prompt)
	}

	tuix.UseEffect(func() func() {
		go func() {
			var activeSubAgent string
			agentOut := make(chan agent.ResponseEvent)
			go func() {
				for {
					setFileDiff(agent.EventManager.ReadFromChannel(agent.FILE_DIFF_CHANNEL).(agent.FileChangeEvent))
				}
			}()

			go func() {
				for {
					ev := agent.EventManager.ReadFromChannel(agent.AGENT_OUTPUT_CHANNEL).(agent.ResponseEvent)
					agentOut <- ev
				}
			}()

			var acc []tuix.Element
			var liveLocal string
			promptIdx := -1
			for {
				select {

				case p := <-promptChan:
					liveLocal = p
					setLivePrompt(p)
					promptIdx = len(acc)
					setLivePromptIdx(promptIdx)
				case ev := <-agentOut:
					msg := ev.Message
					style := tuix.NewStyle().Foreground(tuix.Hex("#c8c8c8"))

					if ev.EventType == agent.Tool {

						if ev.Question != "" {
							setQuestionVisible(true)
							setQuestion(struct {
								question string
								options  []string
							}{question: ev.Question, options: ev.Options})
						}

						if ev.SubAgent {
							if activeSubAgent != ev.SubAgentName {
								activeSubAgent = ev.SubAgentName
								msg = fmt.Sprintf("    \nsubagent:%s\n  └──%s", ev.SubAgentName, msg)
								style = style.Foreground(tuix.Hex("#64c3ff")).Bold(true)
							} else {
								msg = fmt.Sprintf("  └──%s", msg)
								style = style.Foreground(tuix.Hex("#848484"))
							}
						} else {
							msg = fmt.Sprintf("  └──%s", msg)
							style = style.Foreground(tuix.Hex("#848484"))
						}
					} else {
						if liveLocal != "" && promptIdx >= 0 {
							promptEl := view.Prompt(tuix.Props{Values: map[string]any{
								"prompt":  liveLocal,
								"running": false,
							}})
							acc = append(acc[:promptIdx], append([]tuix.Element{promptEl}, acc[promptIdx:]...)...)
							// appending an empty line to reduce cluttering
							acc = append(acc, tuix.Text("", tuix.NewStyle()))
							liveLocal = ""
							promptIdx = -1
							setLivePrompt("")
							setLivePromptIdx(-1)
						}
						setActiveSession(false)

					}

					acc = append(acc, tuix.Box(tuix.Props{Padding: [4]int{0, 2, 0, 2}}, tuix.NewStyle(), tuix.MultilineText(
						msg,
						style,
					)))
				}

				snap := make([]tuix.Element, len(acc))
				copy(snap, acc)
				setOutputs(snap)
			}
		}()
		return func() {}
	}, []any{})

	children := []tuix.Element{view.Banner(tuix.Props{})}

	if activeSession && livePrompt != "" && livePromptIdx >= 0 && livePromptIdx <= len(outputs) {
		children = append(children, outputs[:livePromptIdx]...)
		children = append(children, view.Prompt(tuix.Props{Values: map[string]any{
			"prompt":  livePrompt,
			"running": true,
		}}))
		children = append(children, outputs[livePromptIdx:]...)
	} else {
		children = append(children, outputs...)
	}

	if questionVisible {
		children = append(
			children, tuix.Box(
				tuix.Props{Direction: tuix.Column},
				tuix.NewStyle(),
				tuix.Text("", tuix.NewStyle()),
				view.FileDiff(tuix.Props{Values: map[string]any{"fileDiff": fileDiff}}),
				tuix.Text("", tuix.NewStyle()),
				tuix.Text(question.question, tuix.NewStyle()),
				view.Menu(
					tuix.Props{Values: map[string]any{
						"items":            question.options,
						"setSelectedIndex": setSelectedOption,
						"visible":          questionVisible,
					}},
					func(selected string) {
						setOptionSelected(true)
						setSelectedOption(selected)
					},
				),
			))
	}

	if optionSelected {
		go agent.EventManager.WriteToChannel(agent.AGENT_INPUT_CHANNEL, selectedOption)
		setOptionSelected(false)
		setQuestionVisible(false)
	}

	children = append(children, tuix.Box(
		tuix.Props{Direction: tuix.Row, Padding: [4]int{0, 1, 0, 1}},
		tuix.NewStyle().Border(tuix.Border{
			Top: true, Bottom: true,
			Color: tuix.Hex("#646464"),
		}),
		components.Input(">", "_", true, prompt, func(value string) {
			setPrompt(value)
		}),
	),
	)

	if !menuVisible && activeMenuView != "" {
		setActiveMenuView("")
	}

	if menuVisible {
		children = append(children, MainMenu(
			tuix.Props{Values: map[string]any{
				"activeView":    activeMenuView,
				"setActiveView": setActiveMenuView,
				"prompt":        prompt,
				"submitPrompt":  submitPrompt,
			}},
		),
		)
	}

	children = append(children, view.StatusLine(tuix.Props{
		Values: map[string]any{
			"workspacePath": workspace.AbsToTildePath(props.Get("wd").(string)),
			"running":       activeSession,
		},
	}))

	return tuix.Box(
		tuix.Props{
			Direction: tuix.Column,
			Padding:   [4]int{0, 2, 0, 2},
			Width:     tuix.Grow(1),
		},
		tuix.NewStyle(),
		children...,
	)
}
