package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type OpenRouterProvider struct {
	ProviderId string
	Model      OpenRouterModel
}

type OpenRouterModel string

const (
	GPT_5_2                OpenRouterModel = "openai/gpt-5.2"
	MINIMAX_M2_5           OpenRouterModel = "minimax/minimax-m2.5"
	CLAUDE_SONNET_4_6      OpenRouterModel = "anthropic/claude-sonnet-4.6"
	GPT_5_1_CODEX_MINI     OpenRouterModel = "openai/gpt-5.1-codex-mini"
	KIMI_K_2_5             OpenRouterModel = "moonshotai/kimi-k2.5"
	LLAMA_3_3_70B_INSTRUCT OpenRouterModel = "meta-llama/llama-3.3-70b-instruct"
)

func NewOpenRouterProvider() LLMProvider {
	return &OpenRouterProvider{
		Model: MINIMAX_M2_5,
	}
}

type OpenRouterRequest struct {
	Model               OpenRouterModel        `json:"model,omitempty"`
	Messages            []ChatMessage          `json:"messages"`
	Provider            *ProviderConfig        `json:"provider,omitempty"`
	Temperature         *float64               `json:"temperature,omitempty"`
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
	Tools               []ToolDefinition       `json:"tools,omitempty"`
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
	Role             string             `json:"role"`
	Content          string             `json:"content"`
	Refusal          any                `json:"refusal"`
	Reasoning        string             `json:"reasoning"`
	ReasoningDetails []ReasoningDetails `json:"reasoning_details"`
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
	Content   string `json:"content"`
	Role      string `json:"role"`
	Reasoning string `json:"reasoning"`
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

func (p *OpenRouterProvider) Complete(systemPrompt string, userPrompt ...string) (string, error) {

	fmt.Println("Running OpenRouter api call with: ", p.Model)

	err := godotenv.Load()

	if err != nil {
		fmt.Println("Failed to load env file")
	}

	prompts := []ChatMessage{{Content: systemPrompt, Role: "system"}}

	for _, prompt := range userPrompt {
		prompts = append(prompts, ChatMessage{Content: prompt, Role: "user"})
	}

	requestBody := OpenRouterRequest{
		Model:    p.Model,
		Messages: prompts,
		Stream:   true,
	}

	value, err := json.Marshal(requestBody)

	req, err := http.NewRequest(http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(value))

	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENROUTER_API_KEY")))
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	reader := bufio.NewReader(res.Body)

	finalOutput := ""

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			payload := strings.TrimPrefix(line, "data: ")

			if payload == "[DONE]" {
				break
			}

			var ir OpenRouterResponse

			err := json.Unmarshal([]byte(payload), &ir)

			if err != nil {
				return "", err
			}

			//break out once we reach end of the response
			if len(ir.Choices) == 0 {
				break
			}

			if ir.Choices[0].Delta.Content != "" {
				finalOutput = fmt.Sprintf("%s%s", finalOutput, ir.Choices[0].Delta.Content)
				fmt.Print(ir.Choices[0].Delta.Content)
			} else {
				fmt.Print(ir.Choices[0].Delta.Reasoning)
			}

		}
	}

	return finalOutput, nil
}
