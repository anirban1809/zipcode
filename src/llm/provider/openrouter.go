package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"zipcode/src/tools"

	"github.com/joho/godotenv"
)

type OpenRouterProvider struct {
	ProviderId string
	Model      OpenRouterModel
	Tools      []tools.Tool
}

type OpenRouterModel string

const (
	GPT_5_2                OpenRouterModel = "openai/gpt-5.2"
	MINIMAX_M2_5           OpenRouterModel = "minimax/minimax-m2.5"
	CLAUDE_SONNET_4_6      OpenRouterModel = "anthropic/claude-sonnet-4.6"
	GPT_5_1_CODEX_MINI     OpenRouterModel = "openai/gpt-5.1-codex-mini"
	KIMI_K_2_5             OpenRouterModel = "moonshotai/kimi-k2.5"
	LLAMA_3_3_70B_INSTRUCT OpenRouterModel = "meta-llama/llama-3.3-70b-instruct"
	GLM_4_7                OpenRouterModel = "z-ai/glm-4.7"
	QWEN_3_CODER_FLASH     OpenRouterModel = "qwen/qwen3-coder-flash"
	GPT_5_NANO             OpenRouterModel = "openai/gpt-5-nano"
)

func NewOpenRouterProvider() *OpenRouterProvider {
	return &OpenRouterProvider{
		Model: MINIMAX_M2_5,
		Tools: []tools.Tool{
			tools.BashTool,
			tools.CodeSearchTool,
			tools.FileReadTool,
			tools.FileSearchTool,
			tools.FileWriteTool,
		},
	}
}

type OpenRouterRequest struct {
	Model               OpenRouterModel        `json:"model,omitempty"`
	Messages            []Message              `json:"messages"`
	Provider            *ProviderConfig        `json:"provider,omitempty"`
	Temperature         float64                `json:"temperature,omitempty"`
	TopP                *float64               `json:"top_p,omitempty"`
	FrequencyPenalty    *float64               `json:"frequency_penalty,omitempty"`
	PresencePenalty     *float64               `json:"presence_penalty,omitempty"`
	MaxTokens           *int                   `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int                   `json:"max_completion_tokens,omitempty"`
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

type ChatMessage struct {
	Role       string      `json:"role"`    // system | user | assistant | tool
	Content    interface{} `json:"content"` // string OR []ContentPart
	Name       string      `json:"name,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
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
	Usage             Usage     `json:"usage"`
}
type ReasoningDetails struct {
	Format any    `json:"format"`
	Index  int    `json:"index"`
	Type   string `json:"type"`
	Text   string `json:"text"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls"`
	ToolCallId string     `json:"tool_call_id,omitempty"`
}

type Choices struct {
	Logprobs           any          `json:"logprobs"`
	FinishReason       string       `json:"finish_reason"`
	NativeFinishReason string       `json:"native_finish_reason"`
	Index              int          `json:"index"`
	Message            Message      `json:"message"`
	Delta              MessageDelta `json:"delta"`
}

type MessageDelta struct {
	Content   string     `json:"content"`
	Role      string     `json:"role"`
	Reasoning string     `json:"reasoning"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

type ToolCall struct {
	Type     string           `json:"type"`
	Index    int              `json:"index"`
	ID       string           `json:"id"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
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
type Usage struct {
	PromptTokens            int                     `json:"prompt_tokens"`
	CompletionTokens        int                     `json:"completion_tokens"`
	TotalTokens             int                     `json:"total_tokens"`
	Cost                    float64                 `json:"cost"`
	IsByok                  bool                    `json:"is_byok"`
	PromptTokensDetails     PromptTokensDetails     `json:"prompt_tokens_details"`
	CostDetails             CostDetails             `json:"cost_details"`
	CompletionTokensDetails CompletionTokensDetails `json:"completion_tokens_details"`
}

func (p *OpenRouterProvider) SetModel(model OpenRouterModel, nitro bool) {
	if nitro {
		p.Model = OpenRouterModel(fmt.Sprintf("%s:nitro", model))
		return
	}
	p.Model = model
}

type Conversation struct {
	Tools    []tools.Tool
	Messages []Message
}

func (r *OpenRouterProvider) Chat(prev *Conversation) (*Conversation, error) {
	r.SetModel(GPT_5_1_CODEX_MINI, true)
	value, err := r.Complete(prev)

	if err != nil {
		return nil, err
	}

	prev.Messages = append(prev.Messages, value.Choices[0].Message)

	return prev, nil
}

func (p *OpenRouterProvider) Complete(conversation *Conversation) (OpenRouterResponse, error) {
	// fmt.Println("Running OpenRouter api call with: ", p.Model)

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Failed to load env file")
	}

	var prompts []ChatMessage

	for _, message := range conversation.Messages {
		prompts = append(prompts, ChatMessage{Content: message.Content, Role: message.Role})
	}

	requestBody := OpenRouterRequest{
		Model:    p.Model,
		Messages: conversation.Messages,
		Stream:   false,
		Tools:    p.Tools,
	}

	value, err := json.Marshal(requestBody)

	// fmt.Println(string(value))

	req, err := http.NewRequest(http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(value))

	if err != nil {
		return OpenRouterResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENROUTER_API_KEY")))
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return OpenRouterResponse{}, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return OpenRouterResponse{}, err
	}

	var finalResponse OpenRouterResponse
	err = json.Unmarshal(body, &finalResponse)

	if err != nil {
		return OpenRouterResponse{}, err
	}

	return finalResponse, nil
}
