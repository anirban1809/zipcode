package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type Task struct {
	Prompt   string
	Subs     string
	Result   string
	Running  bool
	Question Question
	Spinner  spinner.Model
}

func CreateTask(prompt string) Task {
	s := spinner.New(spinner.WithSpinner(spinner.Dot))

	return Task{
		Prompt:  prompt,
		Running: false,
		Spinner: s,
	}
}

func (t *Task) AppendSub(value string) {
	t.Subs += value
}

func (t *Task) AppendQuestion(value Question) {
	t.Question = value
}

func (t *Task) UpdateResult(value string) {
	t.Result = value
}

func (t Task) Init() tea.Cmd {
	return t.Spinner.Tick
}

func (t Task) Update(msg tea.Msg) (Task, tea.Cmd) {
	var cmd tea.Cmd
	t.Spinner, cmd = t.Spinner.Update(msg)
	return t, cmd
}

func (t Task) View() string {
	if t.Running {
		return fmt.Sprintf("%s %s", t.Spinner.View(), t.Prompt)
	}
	return fmt.Sprintf("✔ %s", t.Prompt)
}
