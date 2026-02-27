package agent

import "fmt"

type Executor struct{}

type ExecutionResult int

const (
	ExecutionSucceeded ExecutionResult = iota
	ExecutionFailed
	ExecutionCancelled
)

func (e Executor) Execute(step PlanStep) ExecutionResult {
	fmt.Println("Executing: ", step.StepTask)
	return ExecutionSucceeded
}
