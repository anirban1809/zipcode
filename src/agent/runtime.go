package agent

import (
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

	// for r.Status != Idle {
	// 	response, err := r.Planner.StartConversation()

	// 	if err != nil {
	// 		return err
	// 	}

	// 	next, err := r.Executor.ProcessResponse(response)
	// }

	r.Planner.StartConversation()

	// projectType, err := r.Planner.ClassifyProjectType()

	// if err != nil {
	// 	return err
	// }

	// intent, err := r.Planner.ClassifyIntent(prompt, projectType)

	// if err != nil {
	// 	return err
	// }

	// files, err := r.Planner.ResolveScope(intent.SearchIdentifiers)

	// r.Planner.GenerateChanges(prompt, intent, files[0:3])

	// //execution completed, status returns to idle
	// r.Status = Idle
	return nil
}
