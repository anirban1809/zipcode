package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"zipcode/src/config"
	"zipcode/src/llm/prompts"
	llm "zipcode/src/llm/provider"
	"zipcode/src/skills"
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
	Prompt        string
	Executor      *Executor
	Status        RuntimeStatus
	LLM           *llm.OpenRouterProvider
	Workspace     *workspace.Workspace
	Tools         []tools.Tool
	InputTokens   int
	OutputTokens  int
	Conversation  llm.Conversation
	Agent         Agent
	Session       string
	ChildRuntime  bool
	SkillRegistry *skills.SkillRegistry
	SkillWatcher  *skills.Watcher
}

func NewRuntime(workspace *workspace.Workspace) Runtime {
	registry, watcher, _ := skills.Init(
		config.INTERNAL_SKILLS_PATH,
		config.GLOBAL_SKILLS_PATH,
		config.PROJECT_SKILLS_PATH,
		config.SKILLS_STATE_PATH,
	)

	runtime := Runtime{
		Status:    Idle,
		LLM:       llm.NewOpenRouterProvider(),
		Workspace: workspace,
		Executor: NewExecutor(prompts.MainSystemPrompt,
			[]tools.Tool{
				tools.FileWriteTool,
				tools.SubAgentTool,
				tools.InvokeSkillTool,
			},
		),
		SkillRegistry: registry,
		SkillWatcher:  watcher,
	}
	if workspace != nil && workspace.Session != nil {
		runtime.Session = workspace.Session.ID
	}
	runtime.Agent = NewAgent(prompts.MainSystemPrompt, &runtime.Tools, llm.NewOpenRouterProvider())
	runtime.Tools = append(runtime.Tools, tools.FileWriteTool, tools.InvokeSkillTool)
	runtime.loadTools(config.INTERNAL_TOOL_PATH)
	runtime.loadTools(config.EXTERNAL_TOOL_PATH)
	return runtime
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

func (r *Runtime) SetSession(session *workspace.Session) {
	if session == nil {
		return
	}
	r.Session = session.ID
	if r.Workspace != nil {
		r.Workspace.Session = session
	}
	r.Agent.RestoreConversation(session.Messages)
}

func (r *Runtime) persistSessionHistory() {
	if r.Workspace == nil || r.Workspace.Session == nil {
		return
	}
	messages := r.Agent.Conversation.Messages
	if len(messages) > 0 && messages[0].Role == "system" {
		messages = messages[1:]
	}
	r.Workspace.Session.Messages = messages
	_ = r.Workspace.Session.Save()
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
		Status:        Idle,
		Workspace:     r.Workspace,
		Executor:      r.Executor,
		Agent:         childAgent,
		Session:       r.Session,
		ChildRuntime:  true,
		SkillRegistry: r.SkillRegistry,
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

type SkillToolArgs struct {
	SkillName string `json:"skill_name"`
	Args      string `json:"args,omitempty"`
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

// ParseSkillCommand scans the prompt for a /skill-name token (anywhere, not
// just at the start) that resolves to a registered enabled skill. Returns the
// skill name and the rest of the prompt with the token removed; ok=false if
// no enabled skill is referenced.
func (r *Runtime) ParseSkillCommand(prompt string) (name, args string, ok bool) {
	if r.SkillRegistry == nil {
		return "", "", false
	}

	fields := strings.Fields(prompt)
	for i, tok := range fields {
		if !strings.HasPrefix(tok, "/") {
			continue
		}
		candidate := strings.TrimPrefix(tok, "/")
		if candidate == "" {
			continue
		}
		skill, found := r.SkillRegistry.Get(candidate)
		if !found || !skill.Enabled {
			continue
		}
		rest := append([]string{}, fields[:i]...)
		rest = append(rest, fields[i+1:]...)
		return skill.Name, strings.TrimSpace(strings.Join(rest, " ")), true
	}
	return "", "", false
}

func (r *Runtime) IsSkillCommand(prompt string) bool {
	_, _, ok := r.ParseSkillCommand(prompt)
	return ok
}

func (r *Runtime) ExpandSkillCommand(prompt string) string {
	name, args, ok := r.ParseSkillCommand(prompt)
	if !ok {
		return prompt
	}
	skill, found := r.SkillRegistry.Get(name)
	if !found {
		return prompt
	}
	return skills.Resolve(skill.PromptTemplate, args, r.Workspace)
}

func (r *Runtime) skillSummaries() []prompts.SkillSummary {
	if r.SkillRegistry == nil {
		return nil
	}
	enabled := r.SkillRegistry.ListEnabled()
	out := make([]prompts.SkillSummary, 0, len(enabled))
	for _, s := range enabled {
		out = append(out, prompts.SkillSummary{Name: s.Name, Description: s.Description})
	}
	return out
}

func (r *Runtime) refreshSystemPrompt() {
	if r.ChildRuntime {
		return
	}
	prompt := prompts.BuildSystemPrompt(r.skillSummaries())
	r.Agent.SystemPrompt = prompt
	if len(r.Agent.Conversation.Messages) > 0 && r.Agent.Conversation.Messages[0].Role == "system" {
		r.Agent.Conversation.Messages[0].Content = prompt
	}
}

func (r *Runtime) invokeSkill(tool ToolCallResponseData) (llm.Message, *string, error) {
	var args SkillToolArgs
	if err := json.Unmarshal(tool.Arguments, &args); err != nil {
		return llm.Message{
			Role:       "tool",
			ToolCallId: tool.Id,
			Content:    fmt.Sprintf(`{"success":false,"error":"invalid skill args: %s"}`, err.Error()),
		}, nil, nil
	}

	if r.SkillRegistry == nil {
		return llm.Message{
			Role:       "tool",
			ToolCallId: tool.Id,
			Content:    `{"success":false,"error":"skills not available"}`,
		}, nil, nil
	}

	skill, ok := r.SkillRegistry.Get(args.SkillName)
	if !ok || !skill.Enabled {
		return llm.Message{
			Role:       "tool",
			ToolCallId: tool.Id,
			Content:    fmt.Sprintf(`{"success":false,"error":"unknown or disabled skill: %s"}`, args.SkillName),
		}, nil, nil
	}

	resolved := skills.Resolve(skill.PromptTemplate, args.Args, r.Workspace)

	ack := llm.Message{
		Role:       "tool",
		ToolCallId: tool.Id,
		Content:    fmt.Sprintf(`{"success":true,"skill":"%s"}`, skill.Name),
	}
	return ack, &resolved, nil
}

func (r *Runtime) Run(prompt string) (*llm.Message, error) {
	r.Status = Running
	r.Prompt = prompt
	r.refreshSystemPrompt()

	conv, err := r.Agent.RunStep(llm.Message{
		Role:    "user",
		Content: prompt,
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
		var pendingSkillPrompts []string
		var pendingSkillNames []string

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

			case ActionSkill:
				ack, resolved, err := r.invokeSkill(*action.ToolCall)
				if err != nil {
					return nil, err
				}
				messages = append(messages, ack)

				if resolved != nil {
					var args SkillToolArgs
					_ = json.Unmarshal((*action.ToolCall).Arguments, &args)
					pendingSkillPrompts = append(pendingSkillPrompts, *resolved)
					pendingSkillNames = append(pendingSkillNames, args.SkillName)
				}
			}

		}

		for i, body := range pendingSkillPrompts {
			messages = append(messages, llm.Message{
				Role:    "user",
				Content: body,
			})
			r.Executor.SetActiveSkill(pendingSkillNames[i])
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
	r.Conversation.CompletionTokens += r.OutputTokens
	r.Conversation.TotalTokens += r.InputTokens + r.OutputTokens

	r.persistSessionHistory()
	r.Executor.SetActiveSkill("")

	return &conv.Messages[len(conv.Messages)-1], nil
}
