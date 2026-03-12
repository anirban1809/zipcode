package agent

import (
	"encoding/json"
	"errors"
	"strings"
	"zipcode/src/config"
	llm "zipcode/src/llm/provider"
	"zipcode/src/tools"
	"zipcode/src/utils"
)

type Executor struct {
	Events chan string
}

func NewExecutor() Executor {
	return Executor{
		Events: make(chan string),
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

// ----------------------------
// TOOL RESULT REQUEST
// ----------------------------

type ToolResultRequestData struct {
	ToolCallID string `json:"tool_call_id"`
	Role       string `json:"role"`
	Content    string `json:"content"`
}

// ----------------------------
// TOOL CALL RESPONSE
// ----------------------------

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

func (e *Executor) ProcessResponse(response llm.Message) (string, ExecutionResultStatus, error) {
	utils.PrintStruct(response)

	if response.ToolCalls == nil && response.Content != "" {
		var content NormalResponseContent
		err := json.Unmarshal([]byte(response.Content), &content)

		if err != nil {
			return "", ExecutionFailed, err
		}

		e.pushEvent(content.Data.Message)
		return content.Data.Message, ExecutionCompleted, nil
	}

	if response.ToolCalls == nil && strings.Contains(response.Content, "finish") {
		return response.Content, ExecutionCompleted, nil
	}

	if response.Content == "" && len(response.ToolCalls) > 0 {
		tool := ToolCallResponseData{
			Id:        response.ToolCalls[0].ID,
			Name:      response.ToolCalls[0].Function.Name,
			Arguments: json.RawMessage(response.ToolCalls[0].Function.Arguments),
		}
		request, err := e.ProcessToolCall(tool)
		if err != nil {
			return "", ExecutionFailed, err
		}

		requestJson, err := json.Marshal(request)
		return string(requestJson), ExecutionSucceeded, nil
	}

	return "", ExecutionFailed, errors.New("invalid response type")
}

func (e *Executor) pushEvent(value string) {
	if !config.HEADLESS {
		e.Events <- value
	}
}

func (e *Executor) ProcessToolCall(input ToolCallResponseData) (*ToolResultRequestData, error) {
	switch input.Name {
	case "bash_tool":
		var bashInput tools.BashInput
		err := json.Unmarshal(input.Arguments, &bashInput)

		e.pushEvent(bashInput.Message)

		if err != nil {
			return nil, err
		}
		output, err := tools.RunBash(bashInput)
		value, err := json.Marshal(output)

		return &ToolResultRequestData{
			ToolCallID: input.Id,
			Role:       "tool",
			Content:    string(value),
		}, nil

	case "code_search":
		var codeSearchInput tools.CodeSearchInput
		err := json.Unmarshal(input.Arguments, &codeSearchInput)

		e.pushEvent(codeSearchInput.Message)

		if err != nil {
			return nil, err
		}

		output, err := tools.RunCodeSearch(codeSearchInput)

		if err != nil {
			return nil, err
		}

		value, err := json.Marshal(output)

		return &ToolResultRequestData{
			ToolCallID: input.Id,
			Role:       "tool",
			Content:    string(value),
		}, nil

	case "file_search":
		var fileSearchInput tools.FileSearchInput
		err := json.Unmarshal(input.Arguments, &fileSearchInput)

		e.pushEvent(fileSearchInput.Message)

		if err != nil {
			return nil, err
		}

		output, err := tools.RunFileSearch(fileSearchInput)

		if err != nil {
			return nil, err
		}

		value, err := json.Marshal(output)

		return &ToolResultRequestData{
			ToolCallID: input.Id,
			Role:       "tool",
			Content:    string(value),
		}, nil

	case "file_read":
		var fileReadInput tools.FileReadInput
		err := json.Unmarshal(input.Arguments, &fileReadInput)

		e.pushEvent(fileReadInput.Message)

		if err != nil {
			return nil, err
		}

		output, err := tools.RunFileRead(fileReadInput)

		if err != nil {
			return nil, err
		}

		value, err := json.Marshal(output)

		return &ToolResultRequestData{
			ToolCallID: input.Id,
			Role:       "tool",
			Content:    string(value),
		}, nil

	case "file_write":
		var fileWriteInput tools.FileWriteInput
		err := json.Unmarshal(input.Arguments, &fileWriteInput)
		if err != nil {
			return nil, err
		}

		e.pushEvent(fileWriteInput.Message)

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
	return nil, errors.New("invalid tool name")
}
