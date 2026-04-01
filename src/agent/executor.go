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
)

type ResponseEventType int

const (
	Tool ResponseEventType = iota
	Message
)

type ResponseEvent struct {
	Question  string
	Options   []string
	EventType ResponseEventType
	Message   string
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
	EventChannel   chan ResponseEvent
	MessageChannel chan string
}

func NewExecutor() Executor {
	return Executor{
		EventChannel:   make(chan ResponseEvent),
		MessageChannel: make(chan string),
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

type NormalResponseContent struct {
	Type string `json:"type"`
	Data struct {
		Message string `json:"message"`
	} `json:"data"`
}

func (e *Executor) ProcessResponse(response llm.Message) ([]llm.Message, ExecutionResultStatus, error) {
	if config.HEADLESS {
		utils.PrintStruct(response)
	}

	if response.ToolCalls == nil && strings.TrimSpace(response.Content) == "" {
		return []llm.Message{{Role: "user", Content: "Empty response, please retry"}}, ExecutionSucceeded, nil
	}

	if response.ToolCalls == nil && strings.TrimSpace(response.Content) != "" {
		var content NormalResponseContent
		err := json.Unmarshal([]byte(response.Content), &content)

		if err != nil {
			// unmarshalling failed implies that the llm returned a plain string instead
			// of a JSON response. We'll use the string as the executor response
			e.pushEvent(Message, response.Content)
			return nil, ExecutionCompleted, nil
		}

		e.pushEvent(Message, content.Data.Message)
		return nil, ExecutionCompleted, nil
	}

	if len(response.ToolCalls) > 0 {
		results := []llm.Message{}

		for _, toolCall := range response.ToolCalls {
			tool := ToolCallResponseData{
				Id:        toolCall.ID,
				Name:      toolCall.Function.Name,
				Arguments: json.RawMessage(toolCall.Function.Arguments),
			}
			result, err := e.ProcessToolCall(tool)

			if err != nil {
				return nil, ExecutionFailed, err
			}

			results = append(results, llm.Message{
				Role:       result.Role,
				Content:    result.Content,
				ToolCallId: result.ToolCallID,
			})

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
		EventType: eventType,
		Message:   value,
	})
}

func (e *Executor) loadTools(toolsLocation string) ([]string, error) {
	tools := []string{}
	entries, err := os.ReadDir(toolsLocation)

	if err != nil {
		return nil, errors.New("Failed to load tools")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			tools = append(tools, entry.Name())
		}
	}

	return tools, nil
}

func (e *Executor) ProcessToolCall(input ToolCallResponseData) (*ToolResultRequestData, error) {
	switch input.Name {
	case "bash_tool":
		var bashInput tools.BashInput
		err := json.Unmarshal(input.Arguments, &bashInput)

		e.pushEvent(Tool, fmt.Sprintf("%s (%s)", bashInput.Message, bashInput.Command))

		if err != nil {
			return nil, err
		}

		command := fmt.Sprintf("python3 ./src/tools/bash/bash.py --message \"%s\" --command \"%s\"", bashInput.Message, bashInput.Command)
		result, err := tools.RunBashCommand(command, bashInput.WorkingDirectory)

		if err != nil {
			return nil, err
		}

		return &ToolResultRequestData{
			ToolCallID: input.Id,
			Role:       "tool",
			Content:    string(result),
		}, nil

		// case "file_search":
		// 	var fileSearchInput tools.FileSearchInput
		// 	err := json.Unmarshal(input.Arguments, &fileSearchInput)

		// 	e.pushEvent(Tool, fileSearchInput.Message)

		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	output, err := tools.RunFileSearch(fileSearchInput)

		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	value, err := json.Marshal(output)

		// 	return &ToolResultRequestData{
		// 		ToolCallID: input.Id,
		// 		Role:       "tool",
		// 		Content:    string(value),
		// 	}, nil

		// case "file_read":
		// 	var fileReadInput tools.FileReadInput
		// 	err := json.Unmarshal(input.Arguments, &fileReadInput)

		// 	e.pushEvent(Tool, fileReadInput.Message)

		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	output, err := tools.RunFileRead(fileReadInput)

		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	value, err := json.Marshal(output)

		// 	return &ToolResultRequestData{
		// 		ToolCallID: input.Id,
		// 		Role:       "tool",
		// 		Content:    string(value),
		// 	}, nil

		// case "file_write":
		// 	var fileWriteInput tools.FileWriteInput
		// 	err := json.Unmarshal(input.Arguments, &fileWriteInput)
		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	var msg string
		// 	var patches []tools.ParsedDiff

		// 	for _, p := range fileWriteInput.Patches {
		// 		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		// 			A:        difflib.SplitLines(p.Target),
		// 			B:        difflib.SplitLines(p.Content),
		// 			FromFile: fileWriteInput.FilePath,
		// 			ToFile:   fileWriteInput.FilePath,
		// 			Context:  3,
		// 		})

		// 		parsedDiff, _ := tools.ParseUnifiedDiff(diff)
		// 		patches = append(patches, parsedDiff)
		// 	}

		// 	var changeType FileChangeType

		// 	switch fileWriteInput.Operation {
		// 	case "append":
		// 		changeType = FileChange_Append

		// 	case "create":
		// 		changeType = FileChange_Create

		// 	case "patch":
		// 		changeType = FileChange_Patch

		// 	}

		// 	if !config.HEADLESS {
		// 		EventManager.WriteToChannel(FILE_DIFF_CHANNEL, FileChangeEvent{
		// 			FileName:   fileWriteInput.FilePath,
		// 			ChangeType: changeType,
		// 			Content:    fileWriteInput.Content,
		// 			Patches:    patches,
		// 		})

		// 		EventManager.WriteToChannel(AGENT_OUTPUT_CHANNEL, ResponseEvent{
		// 			Question:  "Do you want to make this change?",
		// 			Options:   []string{"Yes", "No", "Yes, and do not ask again for this session"},
		// 			EventType: Tool,
		// 			Message:   fileWriteInput.Message,
		// 		})

		// 		msg = EventManager.ReadFromChannel(AGENT_INPUT_CHANNEL).(string)
		// 	} else {
		// 		msg = "Yes"
		// 	}

		// 	if msg == "Yes" || msg == "Yes, and do not ask again for this session" {
		// 		output, err := tools.RunFileWrite(fileWriteInput)

		// 		if err != nil {
		// 			return nil, err
		// 		}
		// 		value, err := json.Marshal(output)

		// 		return &ToolResultRequestData{
		// 			ToolCallID: input.Id,
		// 			Role:       "tool",
		// 			Content:    string(value),
		// 		}, nil
		// 	}

		// 	return &ToolResultRequestData{
		// 		ToolCallID: input.Id,
		// 		Role:       "tool",
		// 		Content:    string("Action denied by user"),
		// 	}, nil

	}

	return &ToolResultRequestData{
		ToolCallID: input.Id,
		Role:       "tool",
		Content:    fmt.Sprintf(`{"message":"invalid tool name %s, please retry"}`, input.Name),
	}, nil
}
