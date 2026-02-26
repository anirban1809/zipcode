package agent

import llm "zipcode/src/llm/provider"

type RuntimeStatus int

const (
	Initializing RuntimeStatus = iota
	Planning
	Running
	Succeeded
	Failed
	Cancelled
)

type NotImplemented struct{}

type Runtime struct {
	ID           string
	StartTime    string
	Status       RuntimeStatus
	PolicyEngine NotImplemented
	Planner      NotImplemented
	Executor     NotImplemented
	EventBus     NotImplemented
	Store        NotImplemented //optional
	Ctx          NotImplemented
}

func NewRuntime(llmProvider llm.LLMProvider) Runtime {
	return Runtime{}
}
