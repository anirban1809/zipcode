package agent

import (
	"fmt"
	llm "zipcode/src/llm/provider"
	"zipcode/src/workspace"
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
	llm llm.LLMProvider
}

func NewPlanner(workspace *workspace.Workspace) Planner {
	return Planner{
		llm: llm.NewOpenAIProvider(),
	}
}

func CreatePlanStep(stepId int, task string) PlanStep {
	return PlanStep{
		StepId:   stepId,
		StepTask: task,
	}
}

func (p *Planner) CreatePlan(prompt string) Plan {
	response, err := p.llm.Complete("", "Create a plan")

	if err != nil {
		panic(err)
	}

	fmt.Println(response)

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

type StepDecision int

const (
	Allowed StepDecision = iota
	Blocked
	NeedApproval
)

type StepValidationResult struct {
	Valid    bool
	Error    string
	Warning  string
	Decision StepDecision
}

func (p Planner) ValidatePlan(plan *Plan) []StepValidationResult {
	validationResult := []StepValidationResult{}

	for i, step := range plan.Steps {
		if i == 2 {
			validationResult = append(
				validationResult,
				StepValidationResult{
					Valid:    false,
					Error:    "Invariant violation",
					Decision: Blocked,
				},
			)
		}

		validationResult = append(validationResult, p.ValidateStep(&step))
	}
	return validationResult
}

func (p Planner) ValidateStep(step *PlanStep) StepValidationResult {
	return StepValidationResult{Valid: true}
}
