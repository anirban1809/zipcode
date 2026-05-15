package agent

type PlanStepStatus int

const (
	PlanStepPending PlanStepStatus = iota
	PlanStepRunning
	PlanStepCompleted
	PlanStepFailed
)

type PlanStep struct {
	Outline string
	Prompt  string
	Output  string
	Status  PlanStepStatus
}

type Plan struct {
	Title   string
	Steps   []PlanStep
	Current int
	Active  bool
}

type PlanStatusEvent struct {
	Title   string
	Steps   []PlanStep
	Current int
	Active  bool
}

func newPlan(title string, outlines []string) *Plan {
	steps := make([]PlanStep, len(outlines))
	for i, o := range outlines {
		steps[i] = PlanStep{Outline: o, Status: PlanStepPending}
	}
	return &Plan{
		Title:  title,
		Steps:  steps,
		Active: true,
	}
}

func (p *Plan) snapshot() PlanStatusEvent {
	steps := make([]PlanStep, len(p.Steps))
	copy(steps, p.Steps)
	return PlanStatusEvent{
		Title:   p.Title,
		Steps:   steps,
		Current: p.Current,
		Active:  p.Active,
	}
}
