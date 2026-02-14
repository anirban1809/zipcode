package agent

type Runtime struct {
	prompt string
}

func (m *Runtime) Submit(prompt string) {
	m.prompt = prompt
}
