package agent

import (
	"fmt"
	"zipcode/src/workspace"
)

// remove once all types are implemented
type NotImplemented struct{}

// represents the current status of the runtime
type RuntimeStatus int

const (
	Idle RuntimeStatus = iota
	Running
	Cancelled
)

type Runtime struct {
	Prompt    string
	Planner   Planner
	Executor  Executor
	Status    RuntimeStatus
	Workspace *workspace.Workspace
}

func NewRuntime(workspace *workspace.Workspace) Runtime {
	return Runtime{
		Status:    Idle,
		Planner:   NewPlanner(workspace),
		Workspace: workspace,
	}
}

func (r Runtime) Run(prompt string) error {
	r.Status = Running
	r.Prompt = prompt

	intent, err := r.Planner.ClassifyIntent(prompt)

	if err != nil {
		return err
	}

	plan := r.Planner.CreatePlan(r.Prompt, intent, r.Workspace)
	validationReport := r.Planner.ValidatePlan(&plan)

	for i, step := range plan.Steps {
		var status ExecutionResult

		if validationReport[i].Valid {
			status = r.Executor.Execute(step)
			continue
		} else {
			fmt.Println("Plan validation failed, halting operation. Reason:", validationReport[i].Error)
		}

		if status == ExecutionFailed {
			fmt.Println("Execution Failed, exiting operation")
			break
		}
	}

	//execution completed, status returns to idle
	r.Status = Idle
	return nil
}
