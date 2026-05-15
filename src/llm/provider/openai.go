package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"zipcode/src/config"
	"zipcode/src/llm/errors"
	"zipcode/src/tools"
)

type OpenAI struct {
	ProviderId string
	Model      string
	Tools      []tools.Tool
	ApiKey     string
}

func (p OpenAI) Name() ProviderName {
	return OpenAIProvider
}

func (p *OpenAI) SetApiKey(key string) {
	p.ApiKey = key
}

func (p *OpenAI) AuthCheck(key string) AuthResult {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.openai.com/v1/models",
		nil,
	)
	if err != nil {
		return AuthResult{Status: 0, ErrorMessage: err.Error()}
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))

	resp, err := client.Do(req)
	if err != nil {
		return AuthResult{Status: 0, ErrorMessage: err.Error()}
	}
	defer resp.Body.Close()

	result := AuthResult{Status: resp.StatusCode}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		result.ErrorMessage = string(body)
	} else {
		p.ApiKey = key
	}
	return result
}

type openAIRequest struct {
	Model       string       `json:"model"`
	Messages    []Message    `json:"messages"`
	Tools       []tools.Tool `json:"tools,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Temperature float64      `json:"temperature,omitempty"`
}

type openAIResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func (p OpenAI) Complete(request ChatRequest) (ChatResponse, error) {
	body, err := json.Marshal(openAIRequest{
		Model:       config.Cfg.CurrentModel,
		Messages:    request.Messages,
		Tools:       request.Tools,
		MaxTokens:   request.MaxTokens,
		Temperature: request.Temperature,
	})
	if err != nil {
		return ChatResponse{}, err
	}

	res, err := errors.RetryWithBackoff(p, func() (*http.Response, error) {
		req, err := http.NewRequest(
			http.MethodPost,
			"https://api.openai.com/v1/chat/completions",
			bytes.NewReader(body),
		)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.ApiKey))
		return http.DefaultClient.Do(req)
	})
	if err != nil {
		return ChatResponse{}, err
	}

	respBody, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return ChatResponse{}, err
	}

	if res.StatusCode >= 400 {
		return ChatResponse{}, fmt.Errorf(
			"openai: status %d: %s",
			res.StatusCode,
			string(respBody),
		)
	}

	var parsed openAIResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return ChatResponse{}, err
	}

	if len(parsed.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf("openai: no choices in response")
	}

	return ChatResponse{
		ID:    parsed.ID,
		Model: parsed.Model,
		Usage: Usage{
			InputTokens:  parsed.Usage.PromptTokens,
			OutputTokens: parsed.Usage.CompletionTokens,
		},
		Message: Message{
			Role:      parsed.Choices[0].Message.Role,
			Content:   parsed.Choices[0].Message.Content,
			ToolCalls: parsed.Choices[0].Message.ToolCalls,
		},
	}, nil
}

func (p OpenAI) Models() []ModelDescriptor {
	ids := []string{
		"gpt-5.2",
		"gpt-5.5",
		"gpt-5.4",
		"gpt-5.4-nano",
		"gpt-5.3-codex",
		"gpt-5.1-codex-mini",
		"gpt-5-nano",
	}
	descriptors := make([]ModelDescriptor, len(ids))
	for i, id := range ids {
		descriptors[i] = ModelDescriptor{
			ID:           id,
			DisplayName:  id,
			ProviderName: string(OpenAIProvider),
		}
	}
	return descriptors
}

func (p OpenAI) IsQuotaError(resp *http.Response, body []byte) bool {
	if resp.StatusCode != http.StatusTooManyRequests &&
		resp.StatusCode != http.StatusPaymentRequired {
		return false
	}
	var parsed struct {
		Error struct {
			Code string `json:"code"`
			Type string `json:"type"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return false
	}
	return parsed.Error.Code == "insufficient_quota" ||
		parsed.Error.Type == "insufficient_quota"
}
