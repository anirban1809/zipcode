package agent

import (
	"fmt"
	llm "zipcode/src/llm/provider"
)

type Plan struct {
	Steps       []PlanStep
	Validations []StepValidationResult
}

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

func (p Planner) CreatePlan(prompt string) Plan {

	initialStep := 0
	steps := []PlanStep{}

	for i := 0; i < 10; i++ {
		steps = append(steps, CreatePlanStep(initialStep, fmt.Sprintf("Task: %d", initialStep)))
		initialStep++
	}

	return Plan{
		Steps: steps,
	}
}

type StepValidationResult struct {
	Valid         bool
	InvalidReason string
}

func (p Planner) ValidatePlan(plan *Plan) []StepValidationResult {
	validationResult := []StepValidationResult{}

	for i, step := range plan.Steps {

		if i == 2 {
			validationResult = append(validationResult, StepValidationResult{Valid: false})
		}

		validationResult = append(validationResult, p.ValidateStep(&step))
	}

	return validationResult
}

func (p Planner) ValidateStep(step *PlanStep) StepValidationResult {
	return StepValidationResult{Valid: true}
}
