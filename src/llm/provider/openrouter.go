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
	"zipcode/src/utils"

	"github.com/joho/godotenv"
)

type OpenRouterProvider struct {
	ProviderId string
	Model      string
	Tools      []tools.Tool
	ApiKey     string
}

func NewOpenRouterProvider() *OpenRouterProvider {
	return &OpenRouterProvider{}
}

type OpenRouterRequest struct {
	Model               string                 `json:"model,omitempty"`
	Messages            []Message              `json:"messages"`
	Provider            *ProviderConfig        `json:"provider,omitempty"`
	Temperature         float64                `json:"temperature,omitempty"`
	TopP                *float64               `json:"top_p,omitempty"`
	FrequencyPenalty    *float64               `json:"frequency_penalty,omitempty"`
	PresencePenalty     *float64               `json:"presence_penalty,omitempty"`
	MaxTokens           int                    `json:"max_tokens,omitempty"`
	MaxCompletionTokens int                    `json:"max_completion_tokens,omitempty"`
	Stop                []string               `json:"stop,omitempty"`
	Stream              bool                   `json:"stream,omitempty"`
	User                string                 `json:"user,omitempty"`
	SessionID           string                 `json:"session_id,omitempty"`
	Modalities          []string               `json:"modalities,omitempty"`
	Plugins             []PluginConfig         `json:"plugins,omitempty"`
	ToolChoice          interface{}            `json:"tool_choice,omitempty"`
	Tools               []tools.Tool           `json:"tools,omitempty"`
	Extra               map[string]interface{} `json:"extra,omitempty"` // forward compatibility
}

type ContentPart struct {
	Type     string        `json:"type"` // text | image_url | input_text | input_image
	Text     string        `json:"text,omitempty"`
	ImageURL *ImageURLPart `json:"image_url,omitempty"`
}

type ProviderConfig struct {
	AllowFallbacks *bool    `json:"allow_fallbacks,omitempty"`
	Order          []string `json:"order,omitempty"`
}

type PluginConfig struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config,omitempty"`
}

type ToolDefinition struct {
	Type     string       `json:"type"` // typically "function"
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"` // JSON Schema
}

type ImageURLPart struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // auto | low | high
}

type OpenRouterResponse struct {
	ID                string    `json:"id"`
	Provider          string    `json:"provider"`
	Model             string    `json:"model"`
	Object            string    `json:"object"`
	Created           int       `json:"created"`
	Choices           []Choices `json:"choices"`
	SystemFingerprint string    `json:"system_fingerprint"`
	Usage             struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}
type ReasoningDetails struct {
	Format any    `json:"format"`
	Index  int    `json:"index"`
	Type   string `json:"type"`
	Text   string `json:"text"`
}

type Choices struct {
	FinishReason       string `json:"finish_reason"`
	NativeFinishReason string `json:"native_finish_reason"`
	Index              int    `json:"index"`
	Message            struct {
		Content   string     `json:"content"`
		Role      string     `json:"role"`
		ToolCalls []ToolCall `json:"tool_calls"`
	} `json:"message"`
	Delta MessageDelta `json:"delta"`
}

type MessageDelta struct {
	Content   string     `json:"content"`
	Role      string     `json:"role"`
	Reasoning string     `json:"reasoning"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
	AudioTokens  int `json:"audio_tokens"`
}
type CostDetails struct {
	UpstreamInferenceCost            float64 `json:"upstream_inference_cost"`
	UpstreamInferencePromptCost      float64 `json:"upstream_inference_prompt_cost"`
	UpstreamInferenceCompletionsCost float64 `json:"upstream_inference_completions_cost"`
}
type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
	AudioTokens     int `json:"audio_tokens"`
}

func (p *OpenRouterProvider) SetModel(model string, nitro bool) {
	if nitro {
		p.Model = fmt.Sprintf("%s:nitro", model)
		return
	}
	p.Model = model
}

func (p *OpenRouterProvider) AuthCheck(key string) AuthResult {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(
		http.MethodGet,
		"https://openrouter.ai/api/v1/key",
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

func (p OpenRouterProvider) IsQuotaError(
	resp *http.Response,
	body []byte,
) bool {
	// OpenRouter uses 402 Payment Required for insufficient credits; 429 is rate-limit only.
	return resp.StatusCode == http.StatusPaymentRequired
}

func (p OpenRouterProvider) Models() []ModelDescriptor {
	ids := []string{
		"openai/gpt-5.2",
		"openai/gpt-5.5",
		"minimax/minimax-m2.5",
		"minimax/minimax-m2.7",
		"anthropic/claude-sonnet-4.6",
		"anthropic/claude-haiku-4.5",
		"openai/gpt-5.1-codex-mini",
		"moonshotai/kimi-k2.5",
		"meta-llama/llama-3.3-70b-instruct",
		"z-ai/glm-4.7",
		"qwen/qwen3-coder-flash",
		"openai/gpt-5-nano",
		"z-ai/glm-5",
		"openai/gpt-5.4-nano",
		"deepseek/deepseek-v3.2",
		"openai/gpt-5.4",
		"openai/gpt-5.3-codex",
		"z-ai/glm-5v-turbo",
	}
	descriptors := make([]ModelDescriptor, len(ids))
	for i, id := range ids {
		descriptors[i] = ModelDescriptor{
			ID:           id,
			DisplayName:  id,
			ProviderName: string(OpenRouterAPIProvider),
		}
	}
	return descriptors
}

func (p OpenRouterProvider) Name() ProviderName {
	return OpenRouterAPIProvider
}

func (p *OpenRouterProvider) SetApiKey(key string) {
	p.ApiKey = key
}

func (p *OpenRouterProvider) Complete(
	request ChatRequest,
) (ChatResponse, error) {
	godotenv.Load()

	retry := true
	var finalResponse OpenRouterResponse

	for retry {
		requestBody := OpenRouterRequest{
			Model:               config.Cfg.CurrentModel,
			Messages:            request.Messages,
			Stream:              false,
			Tools:               request.Tools,
			MaxTokens:           8192,
			MaxCompletionTokens: 2048,
		}

		value, err := json.Marshal(requestBody)
		if err != nil {
			return ChatResponse{}, err
		}

		// debug code
		utils.LogValue(request.Messages[len(request.Messages)-1])
		// debug code

		res, err := errors.RetryWithBackoff(p, func() (*http.Response, error) {
			req, err := http.NewRequest(
				http.MethodPost,
				"https://openrouter.ai/api/v1/chat/completions",
				bytes.NewReader(value),
			)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(
				"Authorization",
				fmt.Sprintf("Bearer %s", p.ApiKey),
			)
			return http.DefaultClient.Do(req)
		})
		if err != nil {
			return ChatResponse{}, err
		}

		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return ChatResponse{}, err
		}

		if err := json.Unmarshal(body, &finalResponse); err != nil {
			return ChatResponse{}, err
		}

		if len(finalResponse.Choices) > 0 {
			retry = false
		} else {
			fmt.Println("retrying", string(body))
		}
	}

	var chatResponse ChatResponse
	chatResponse.Model = finalResponse.Model
	chatResponse.ID = finalResponse.ID
	chatResponse.Usage.InputTokens = finalResponse.Usage.PromptTokens
	chatResponse.Usage.OutputTokens = finalResponse.Usage.CompletionTokens
	chatResponse.Message.Role = finalResponse.Choices[0].Message.Role
	chatResponse.Message.Content = finalResponse.Choices[0].Message.Content
	chatResponse.Message.ToolCalls = finalResponse.Choices[0].Message.ToolCalls

	return chatResponse, nil
}
