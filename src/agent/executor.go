package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"zipcode/src/config"
	llm "zipcode/src/llm/provider"
	"zipcode/src/tools"
	"zipcode/src/utils"

	"github.com/pmezard/go-difflib/difflib"
)

type ResponseEventType int

const (
	Tool ResponseEventType = iota
	Message
)

type ResponseEvent struct {
	Question     string
	Options      []string
	EventType    ResponseEventType
	Message      string
	SubAgent     bool
	SubAgentName string
	SkillName    string
}

type FileChangeType int

const (
	FileChange_Create FileChangeType = iota
	FileChange_Append
	FileChange_Patch
)

type FileChangeEvent struct {
	FileName   string
	ChangeType FileChangeType
	Content    string
	Patches    []tools.ParsedDiff
}

type Executor struct {
	EventChannel    chan ResponseEvent
	MessageChannel  chan string
	SystemPrompt    string
	Tools           []tools.Tool
	SubAgentRunning bool
	SubAgent        string
	ActiveSkill     string
}

func (e *Executor) IsSubagentTool(name string) bool {
	return strings.HasPrefix(name, "subagent")
}

func (e *Executor) IsSkillTool(name string) bool {
	return name == "invoke_skill"
}

func NewExecutor(systemPrompt string, tools []tools.Tool) *Executor {
	return &Executor{
		EventChannel:   make(chan ResponseEvent),
		MessageChannel: make(chan string),
		SystemPrompt:   systemPrompt,
	}
}

type ExecutionResultStatus int

const (
	ExecutionSucceeded ExecutionResultStatus = iota
	ExecutionFailed
	ExecutionCancelled
	ExecutionCompleted
)

type RequestType string

const (
	RequestTask       RequestType = "task"
	RequestToolResult RequestType = "tool_result"
	RequestMessage    RequestType = "message"
)

type ResponseType string

const (
	ResponseToolCall ResponseType = "tool_call"
	ResponseMessage  ResponseType = "message"
	ResponseFinish   ResponseType = "finish"
)

type ToolResultRequestData struct {
	ToolCallID string `json:"tool_call_id"`
	Role       string `json:"role"`
	Content    string `json:"content"`
}

type ToolCallResponseData struct {
	Id        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ExecutionActionType string

const (
	ActionMessage  ExecutionActionType = "message"
	ActionToolCall ExecutionActionType = "tool_call"
	ActionSubagent ExecutionActionType = "subagent"
	ActionSkill    ExecutionActionType = "skill"
	ActionComplete ExecutionActionType = "complete"
)

type ExecutionAction struct {
	Type     ExecutionActionType
	Message  *llm.Message
	ToolCall *ToolCallResponseData
}

func (e *Executor) SetSubAgentModeOn(mode bool, name string) {
	e.SubAgent = name
	e.SubAgentRunning = mode
}

func (e *Executor) SetActiveSkill(name string) {
	e.ActiveSkill = name
}

func (e *Executor) ProcessResponse(response llm.Message) ([]ExecutionAction, ExecutionResultStatus, error) {
	if config.HEADLESS {
		utils.PrintStruct(response)
	}

	if response.ToolCalls == nil && strings.TrimSpace(response.Content) == "" {
		return []ExecutionAction{
			{Type: ActionMessage, Message: &llm.Message{Role: "user", Content: "retry"}},
		}, ExecutionSucceeded, nil
	}

	if response.ToolCalls == nil && strings.TrimSpace(response.Content) != "" {
		if !e.SubAgentRunning {
			e.pushEvent(Message, response.Content)
		}
		return nil, ExecutionCompleted, nil
	}

	if len(response.ToolCalls) > 0 {
		results := []ExecutionAction{}

		for _, toolCall := range response.ToolCalls {
			tool := ToolCallResponseData{
				Id:        toolCall.ID,
				Name:      toolCall.Function.Name,
				Arguments: json.RawMessage(toolCall.Function.Arguments),
			}

			actionType := ActionToolCall

			if e.IsSubagentTool(tool.Name) {
				actionType = ActionSubagent
			}

			if e.IsSkillTool(tool.Name) {
				actionType = ActionSkill
			}

			results = append(results, ExecutionAction{Type: actionType, ToolCall: &tool})

		}

		return results, ExecutionSucceeded, nil
	}

	return nil, ExecutionFailed, errors.New("invalid response type")
}

func (e *Executor) pushEvent(eventType ResponseEventType, value string) {
	if config.HEADLESS {
		return
	}

	EventManager.WriteToChannel(AGENT_OUTPUT_CHANNEL, ResponseEvent{
		EventType:    eventType,
		Message:      value,
		SubAgent:     e.SubAgentRunning,
		SubAgentName: e.SubAgent,
		SkillName:    e.ActiveSkill,
	})
}

func GetTool(path string, toolname string) (tools.Tool, error) {
	name := strings.ReplaceAll(toolname, "_tool", "")
	content, err := os.ReadFile(fmt.Sprintf("%s/%s/%s.json", path, name, name))

	if err != nil {
		return tools.Tool{}, errors.New("failed to read tool manifest")
	}

	var tool tools.Tool
	err = json.Unmarshal([]byte(content), &tool)

	if err != nil {
		return tools.Tool{}, errors.New("invalid tool manifest")
	}

	return tool, nil
}

func (e *Executor) GetToolCallCommand(input ToolCallResponseData) (string, error) {
	internaltool, err1 := GetTool(config.INTERNAL_TOOL_PATH, input.Name)
	externaltool, err2 := GetTool(config.EXTERNAL_TOOL_PATH, input.Name)
	var toolPath string
	var tool tools.Tool

	if err1 != nil {
		tool = externaltool
		toolPath = config.EXTERNAL_TOOL_PATH
	} else if err2 != nil {
		tool = internaltool
		toolPath = config.INTERNAL_TOOL_PATH
	}

	if err1 != nil && err2 != nil {
		return "", errors.New("failed to get tool")
	}

	command := fmt.Sprintf("python3 %s/%s/%s.py", toolPath, input.Name, input.Name)

	for _, param := range tool.Function.Parameters.Required {
		var args map[string]any

		if err := json.Unmarshal(input.Arguments, &args); err != nil {
			fmt.Println("Error:", err)
		}

		command = fmt.Sprintf("%s --%s \"%s\"", command, param, strings.ReplaceAll(args[param].(string), "\"", "\\\""))
	}
	return command, nil
}

func (e *Executor) ProcessToolCall(input ToolCallResponseData) (*ToolResultRequestData, error) {
	switch input.Name {
	default:
		command, err := e.GetToolCallCommand(input)
		if err != nil {
			return &ToolResultRequestData{
				ToolCallID: input.Id,
				Role:       "tool",
				Content:    fmt.Sprintf("unknown tool: %s", input.Name),
			}, nil
		}

		utils.Log(command)

		var args map[string]any

		if err := json.Unmarshal(input.Arguments, &args); err != nil {
			return nil, err
		}

		e.pushEvent(Tool, args["message"].(string))

		result, err := tools.RunBashCommand(command)
		utils.Log(result)

		if err != nil {
			utils.Log(err.Error())
			return nil, err
		}

		return &ToolResultRequestData{
			ToolCallID: input.Id,
			Role:       "tool",
			Content:    string(result),
		}, nil

	case "file_write":
		var fileWriteInput tools.FileWriteInput
		err := json.Unmarshal(input.Arguments, &fileWriteInput)
		if err != nil {
			return nil, err
		}

		var msg string
		var patches []tools.ParsedDiff

		for _, p := range fileWriteInput.Patches {
			diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(p.Target),
				B:        difflib.SplitLines(p.Content),
				FromFile: fileWriteInput.FilePath,
				ToFile:   fileWriteInput.FilePath,
				Context:  3,
			})

			parsedDiff, _ := tools.ParseUnifiedDiff(diff)
			patches = append(patches, parsedDiff)
		}

		var changeType FileChangeType

		switch fileWriteInput.Operation {
		case "append":
			changeType = FileChange_Append

		case "create":
			changeType = FileChange_Create

		case "patch":
			changeType = FileChange_Patch

		}

		if !config.HEADLESS {
			EventManager.WriteToChannel(FILE_DIFF_CHANNEL, FileChangeEvent{
				FileName:   fileWriteInput.FilePath,
				ChangeType: changeType,
				Content:    fileWriteInput.Content,
				Patches:    patches,
			})

			EventManager.WriteToChannel(AGENT_OUTPUT_CHANNEL, ResponseEvent{
				Question:  "Do you want to make this change?",
				Options:   []string{"Yes", "No"},
				EventType: Tool,
				Message:   fileWriteInput.Message,
			})

			msg = EventManager.ReadFromChannel(AGENT_INPUT_CHANNEL).(string)
		} else {
			msg = "Yes"
		}

		if msg == "Yes" || msg == "Yes, and do not ask again for this session" {
			output, err := tools.RunFileWrite(fileWriteInput)

			if err != nil {
				return nil, err
			}
			value, err := json.Marshal(output)

			return &ToolResultRequestData{
				ToolCallID: input.Id,
				Role:       "tool",
				Content:    string(value),
			}, nil
		}

		return &ToolResultRequestData{
			ToolCallID: input.Id,
			Role:       "tool",
			Content:    "denied",
		}, nil

	}

}
