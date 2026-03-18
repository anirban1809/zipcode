package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Question struct {
	Question       string
	Options        []string
	Cursor         int
	SelectedOption int
	Selected       bool
	Visible        bool
}

func CreateQuestion(question string, options []string) Question {
	return Question{
		Question: question,
		Options:  options,
		Selected: false,
		Cursor:   0,
	}
}

func (a Question) Init() tea.Cmd {
	return nil
}

func (a Question) Update(msg tea.Msg) (Question, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			if a.Cursor < len(a.Options)-1 {
				a.Cursor++
			}

		case "up":
			if a.Cursor > 0 {
				a.Cursor--
			}

		case "enter":
			a.SelectedOption = a.Cursor
			a.Selected = true
		}
	}

	return a, nil
}

func (a Question) GetSelectedItem() string {
	return a.Options[a.SelectedOption]
}

func (a Question) View() string {
	if !a.Visible {
		return ""
	}

	question := a.Question
	optionText := ""

	for i, option := range a.Options {
		if i == a.Cursor {
			optionText += fmt.Sprintf("-> %s\n", option)
			continue
		}
		optionText += fmt.Sprintf(" %s\n", option)
	}

	return fmt.Sprintf("\n%s\n%s", question, optionText)
}
