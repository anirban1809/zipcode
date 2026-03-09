package agent

import (
	"encoding/json"
	"errors"
	llm "zipcode/src/llm/provider"
	"zipcode/src/tools"
)

type Executor struct{}

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

// Root request envelope
type Request struct {
	Type RequestType `json:"type"`
	Data any         `json:"data"`
}

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
// MESSAGE REQUEST
// ----------------------------

type MessageRequestData struct {
	Message string `json:"message"`
}

// ============================
// RESPONSE TYPES
// ============================

// Root response envelope
type Response struct {
	Type ResponseType    `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ----------------------------
// TOOL CALL RESPONSE
// ----------------------------

type ToolCallResponseData struct {
	Id        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ----------------------------
// MESSAGE RESPONSE
// ----------------------------

type MessageResponseData struct {
	Message string `json:"message"`
}

// ----------------------------
// FINISH RESPONSE
// ----------------------------

type FinishResponseData struct {
	Message string `json:"message"`
}

type ToolResultData struct {
	ToolName string          `json:"tool_name"`
	Result   json.RawMessage `json:"result"`
}

type ToolInputData struct {
	ToolName string `json:"tool_name"`
}

func (e *Executor) ProcessResponse(response llm.Message) (string, ExecutionResultStatus, error) {
	// switch llmResponse.Type {
	// case ResponseToolCall:
	// 	var tool ToolCallResponseData
	// 	json.Unmarshal(llmResponse.Data, &tool)
	// 	request, err := e.ProcessToolCall(tool)
	// 	if err != nil {
	// 		return "", ExecutionFailed, err
	// 	}

	// 	requestJson, err := json.Marshal(request)
	// 	return string(requestJson), ExecutionSucceeded, nil

	// case ResponseFinish:
	// 	var response FinishResponseData
	// 	err := json.Unmarshal(llmResponse.Data, &response)
	// 	if err != nil {
	// 		return "", ExecutionFailed, err
	// 	}

	// 	return response.Message, ExecutionCompleted, nil

	// case ResponseMessage:
	// 	var response MessageResponseData
	// 	err := json.Unmarshal(llmResponse.Data, &response)
	// 	if err != nil {
	// 		return "", ExecutionFailed, err
	// 	}

	// 	return response.Message, ExecutionSucceeded, nil
	// }

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

func (e *Executor) ProcessToolCall(input ToolCallResponseData) (*ToolResultRequestData, error) {
	switch input.Name {
	case "bash_tool":
		var bashInput tools.BashInput
		err := json.Unmarshal(input.Arguments, &bashInput)
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
