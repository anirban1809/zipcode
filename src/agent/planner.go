package agent

import (
	"fmt"
	llm "zipcode/src/llm/provider"
)

type PlanStep struct {
	StepId   int
	StepTask string
}

type Planner struct {
	LLMProvider llm.LLMProvider
}

func CreatePlanStep(stepId int, task string) PlanStep {
	return PlanStep{
		StepId:   stepId,
		StepTask: task,
	}
}

func (p Planner) Plan(prompt string) []PlanStep {

	initialStep := 0
	steps := []PlanStep{}

	for i := 0; i < 10; i++ {
		steps = append(steps, CreatePlanStep(initialStep, fmt.Sprintf("Task: %d", initialStep)))
		initialStep++
	}

	return steps
}
