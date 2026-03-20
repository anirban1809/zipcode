package agent

import (
	"encoding/json"
	"zipcode/src/llm/prompts"
	llm "zipcode/src/llm/provider"
	"zipcode/src/tools"
	"zipcode/src/workspace"
)

// Represents the current status of the runtime
type RuntimeStatus int

const (
	Idle RuntimeStatus = iota
	Running
	Cancelled
)

type RuntimeEvent string

type Runtime struct {
	Prompt       string
	Executor     Executor
	Status       RuntimeStatus
	LLM          *llm.OpenRouterProvider
	Workspace    *workspace.Workspace
	Tools        []tools.Tool
	InputTokens  int
	OutputTokens int
	Conversation llm.Conversation
}

func NewRuntime(workspace *workspace.Workspace) Runtime {
	return Runtime{
		Status:    Idle,
		LLM:       llm.NewOpenRouterProvider(),
		Workspace: workspace,
		Executor:  NewExecutor(),
	}
}

type TaskRequest struct {
	Type string          `json:"type"`
	Data TaskRequestData `json:"data"`
}

type TaskRequestData struct {
	Objective string `json:"objective"`
	Workspace string `json:"workspace,omitempty"`
	Context   string `json:"context,omitempty"`
}

func (r Runtime) GetExecutorEventChannel() chan ResponseEvent {
	return r.Executor.EventChannel
}

func (r Runtime) GetExecutorMessageChannel() chan string {
	return r.Executor.MessageChannel
}

func (r *Runtime) Run(prompt string) error {

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

	var conv *llm.Conversation

	if len(r.Conversation.Messages) == 0 {
		initialConversation := llm.Conversation{
			Messages: []llm.Message{
				{
					Content: prompts.MainSystemPrompt,
					Role:    "system",
				},
				{
					Content: string(taskRequestString),
					Role:    "user",
				},
			},
			Tools: []tools.Tool{
				tools.BashTool,
				tools.CodeSearchTool,
				tools.FileReadTool,
				tools.FileSearchTool,
				tools.FileWriteTool,
			},
		}

		conv, err = r.LLM.Chat(&initialConversation)
	} else {

		r.Conversation.Messages = append(r.Conversation.Messages, llm.Message{
			Content: string(taskRequestString),
			Role:    "user",
		})

		conv, err = r.LLM.Chat(&r.Conversation)
	}

	if err != nil {
		return err
	}

	for r.Status != Idle {
		lastResponseIndex := len(conv.Messages) - 1
		lastResponse := conv.Messages[lastResponseIndex]
		next, status, err := r.Executor.ProcessResponse(lastResponse)

		// utils.Log(next + "\n")

		if err != nil {
			return err
		}

		if status == ExecutionCompleted {
			r.Status = Idle
			break
		}

		var message llm.Message
		err = json.Unmarshal([]byte(next), &message)

		if err != nil {
			return err
		}

		conv.Messages = append(conv.Messages, message)
		conv, err = r.LLM.Chat(conv)

		r.InputTokens += conv.PromptTokens
		r.OutputTokens += conv.CompletionTokens
	}

	r.Conversation.Messages = append(r.Conversation.Messages, conv.Messages...)
	r.Conversation.PromptTokens += r.InputTokens
	r.Conversation.CompletionTokens += r.Conversation.CompletionTokens
	r.Conversation.TotalTokens += r.Conversation.TotalTokens

	return nil
}
