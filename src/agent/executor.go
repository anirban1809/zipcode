package agent

import "encoding/json"

type Executor struct{}

type ExecutionResultStatus int

const (
	ExecutionSucceeded ExecutionResultStatus = iota
	ExecutionFailed
	ExecutionCancelled
	ExecutionCompleted
)

type ExecutionOutput interface {
	string | int
}

type ResponseType string

const (
	TYPE_TOOL_CALL ResponseType = "tool_call"
	TYPE_MESSAGE   ResponseType = "message"
	TYPE_FINISH    ResponseType = "finish"
)

type LLMResponse struct {
	Type ResponseType    `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (e *Executor) ProcessResponse(response LLMResponse) (any, ExecutionResultStatus, error) {
	return nil, 0, nil
}
