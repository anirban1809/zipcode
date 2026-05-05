package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"zipcode/src/config"
	"zipcode/src/llm/prompts"
	llm "zipcode/src/llm/provider"
	"zipcode/src/tools"
	"zipcode/src/utils"
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
	Executor     *Executor
	Status       RuntimeStatus
	LLM          *llm.OpenRouterProvider
	Workspace    *workspace.Workspace
	Tools        []tools.Tool
	InputTokens  int
	OutputTokens int
	Conversation llm.Conversation
	Agent        Agent
	Session      string
	ChildRuntime bool
}

func NewRuntime(workspace *workspace.Workspace) Runtime {
	runtime := Runtime{
		Status:    Idle,
		LLM:       llm.NewOpenRouterProvider(),
		Workspace: workspace,
		Executor: NewExecutor(prompts.MainSystemPrompt,
			[]tools.Tool{
				tools.FileWriteTool,
				tools.SubAgentTool,
			},
		),
	}
	runtime.Agent = NewAgent(prompts.MainSystemPrompt, &runtime.Tools, llm.NewOpenRouterProvider())
	runtime.Tools = append(runtime.Tools, tools.FileWriteTool)
	runtime.loadTools(config.INTERNAL_TOOL_PATH)
	runtime.loadTools(config.EXTERNAL_TOOL_PATH)
	return runtime
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

func (r *Runtime) SetModel(model string) {
	r.LLM.SetModel(model, false)
}

type SubAgentRequest struct {
	AgentName   string
	AgentPrompt string
}

type SubAgent struct {
	Name             string   `json:"name"`
	ShortDescription string   `json:"short_description"`
	SystemPrompt     string   `json:"system_prompt"`
	AllowedTools     []string `json:"allowed_tools"`
}

func GetToolsforSubAgent(toolNames []string) ([]tools.Tool, error) {
	allowedTools := []tools.Tool{}

	for _, toolName := range toolNames {
		toolManifest, err := GetTool(config.INTERNAL_TOOL_PATH, toolName)

		if err != nil {
			return nil, err
		}

		allowedTools = append(allowedTools, toolManifest)
	}

	return allowedTools, nil
}

func (r *Runtime) NewChildRuntime(agentName string) (*Runtime, error) {

	content, err := os.ReadFile(fmt.Sprintf("%s/%s.json", config.INTERNAL_SUBAGENTS_PATH, agentName))
	if err != nil {
		return nil, err
	}

	var subAgentDefinition SubAgent
	err = json.Unmarshal(content, &subAgentDefinition)

	if err != nil {
		return nil, err
	}

	tools, err := GetToolsforSubAgent(subAgentDefinition.AllowedTools)

	childAgent := NewAgent(
		subAgentDefinition.SystemPrompt, &tools, llm.NewOpenRouterProvider(),
	)

	return &Runtime{
		Status:       Idle,
		Workspace:    r.Workspace,
		Executor:     r.Executor,
		Agent:        childAgent,
		Session:      r.Session,
		ChildRuntime: true,
	}, nil
}

func (r *Runtime) loadTools(path string) error {
	entries, err := os.ReadDir(path)

	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			content, err := os.ReadFile(fmt.Sprintf("%s/%s/%s.json", path, entry.Name(), entry.Name()))
			if err != nil {
				fmt.Println("Error reading file:", err)
				continue
			}

			var tool tools.Tool
			err = json.Unmarshal([]byte(content), &tool)

			r.Tools = append(r.Tools, tool)
		}
	}

	return nil
}

type SubagentToolArgs struct {
	AgentName string `json:"agent"`
	Task      string `json:"task"`
	Context   string `json:"context,omitempty"`
}

func (r *Runtime) InvokeSubAgent(tool ToolCallResponseData) (llm.Message, error) {
	var args SubagentToolArgs
	if err := json.Unmarshal(tool.Arguments, &args); err != nil {
		return llm.Message{
			Role:       "tool",
			ToolCallId: tool.Id,
			Content:    fmt.Sprintf(`{"success":false,"error":"invalid subagent args: %s"}`, err.Error()),
		}, nil
	}

	childRuntime, err := r.NewChildRuntime(args.AgentName)
	if err != nil {
		return llm.Message{
			Role:       "tool",
			ToolCallId: tool.Id,
			Content:    fmt.Sprintf(`{"success":false,"error":"failed to create subagent: %s"}`, err.Error()),
		}, nil
	}

	r.Executor.SetSubAgentModeOn(true, args.AgentName)
	output, err := childRuntime.Run(args.Task)
	r.Executor.SetSubAgentModeOn(false, "")
	if err != nil {
		return llm.Message{
			Role:       "tool",
			ToolCallId: tool.Id,
			Content:    fmt.Sprintf(`{"success":false,"error":"subagent failed: %s"}`, err.Error()),
		}, nil
	}

	result := map[string]any{
		"success":    true,
		"agent_type": args.AgentName,
		"output":     output,
	}

	content, _ := json.Marshal(result)

	return llm.Message{
		Role:       "tool",
		ToolCallId: tool.Id,
		Content:    string(content),
	}, nil
}

func (r *Runtime) Run(prompt string) (*llm.Message, error) {
	r.Status = Running
	r.Prompt = prompt

	taskRequest := TaskRequest{
		Type: "task",
		Data: TaskRequestData{
			Objective: prompt,
			Workspace: r.Workspace.RootPath,
		},
	}

	userPrompt, err := json.Marshal(taskRequest)

	if err != nil {
		return nil, err
	}

	var conv *llm.Conversation

	conv, err = r.Agent.RunStep(llm.Message{
		Role:    "user",
		Content: string(userPrompt),
	})

	if err != nil {
		return nil, err
	}

	for r.Status != Idle {
		lastResponseIndex := len(conv.Messages) - 1
		lastResponse := conv.Messages[lastResponseIndex]
		actions, status, err := r.Executor.ProcessResponse(lastResponse)

		utils.Log(lastResponse.Content)

		if err != nil {
			return nil, err
		}

		if status == ExecutionCompleted {
			r.Status = Idle
			break
		}

		messages := []llm.Message{}

		for _, action := range actions {
			switch action.Type {
			case ActionToolCall:

				result, err := r.Executor.ProcessToolCall(*action.ToolCall)
				if err != nil {
					return nil, err
				}

				messages = append(messages, llm.Message{
					Role:       result.Role,
					Content:    result.Content,
					ToolCallId: result.ToolCallID,
				})

			case ActionSubagent:
				var args SubagentToolArgs
				err := json.Unmarshal((*action.ToolCall).Arguments, &args)

				if err != nil {
					return nil, err
				}

				result, err := r.InvokeSubAgent(*action.ToolCall)
				if err != nil {
					return nil, err
				}

				messages = append(messages, result)
			}

		}

		conv, err := r.Agent.RunStep(messages...)

		if err != nil {
			return nil, err
		}

		r.InputTokens += conv.PromptTokens
		r.OutputTokens += conv.CompletionTokens
	}

	r.Conversation.Messages = append(r.Conversation.Messages, conv.Messages...)
	r.Conversation.PromptTokens += r.InputTokens
	r.Conversation.CompletionTokens += r.Conversation.CompletionTokens
	r.Conversation.TotalTokens += r.Conversation.TotalTokens

	return &conv.Messages[len(conv.Messages)-1], nil
}
