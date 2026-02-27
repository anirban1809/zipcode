package agent

import "fmt"

// remove once all types are implemented
type NotImplemented struct{}

type Runtime struct {
	Prompt   string
	Planner  Planner
	Executor Executor
}

func NewRuntime() Runtime {
	return Runtime{}
}

func (r Runtime) Run(prompt string) {
	r.Prompt = prompt
	steps := r.Planner.Plan(r.Prompt)

	for _, step := range steps {
		status := r.Executor.Execute(step)

		if status == ExecutionFailed {
			fmt.Println("Execution Failed, exiting operation")
			break
		}
	}
}
