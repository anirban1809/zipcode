package agent

import (
	"encoding/json"
	"fmt"
	"zipcode/src/llm/prompts"
	llm "zipcode/src/llm/provider"
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
	Executor  Executor
	Status    RuntimeStatus
	LLM       *llm.OpenRouterProvider
	Workspace *workspace.Workspace
}

func NewRuntime(workspace *workspace.Workspace) Runtime {
	return Runtime{
		Status:    Idle,
		LLM:       llm.NewOpenRouterProvider(),
		Workspace: workspace,
	}
}

/*
{
  "type": "task",
  "data": {
    "objective": "<user task description>",
    "workspace": "<workspace path>",
    "context": "<optional context>"
  }
}
*/

type TaskRequest struct {
	Type string          `json:"type"`
	Data TaskRequestData `json:"data"`
}

type TaskRequestData struct {
	Objective string `json:"objective"`
	Workspace string `json:"workspace,omitempty"`
	Context   string `json:"context,omitempty"`
}

func (r Runtime) Run(prompt string) error {
	r.Status = Running
	r.Prompt = prompt

	taskRequest := TaskRequest{
		Type: "task",
		Data: TaskRequestData{
			Objective: prompt,
			Workspace: r.Workspace.RootPath,
		},
	}

	taskRequestString, err := json.Marshal(taskRequest)

	if err != nil {
		return err
	}

	initialConversation := llm.Conversation{
		Messages: []llm.ConversationMessage{
			{
				Content: prompts.MainSystemPrompt,
				Role:    "system",
			},
			{
				Content: string(taskRequestString),
				Role:    "user",
			},
		},
	}

	conv, err := r.LLM.Chat(&initialConversation)

	if err != nil {
		return err
	}

	for r.Status != Idle {
		lastResponseIndex := len(conv.Messages) - 1
		lastResponse := conv.Messages[lastResponseIndex].Content
		next, status, err := r.Executor.ProcessResponse(lastResponse)

		fmt.Println(next)

		if err != nil {
			return err
		}

		if status == ExecutionCompleted {
			r.Status = Idle
			break
		}

		conv.Messages = append(conv.Messages, llm.ConversationMessage{Role: "user", Content: next})
		conv, err = r.LLM.Chat(conv)
	}

	return nil
}
