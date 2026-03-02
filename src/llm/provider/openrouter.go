package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type OpenRouterProvider struct {
	ProviderId string
	Model      OpenRouterModel
}

type OpenRouterModel string

const (
	GPT_5_2      OpenRouterModel = "openai/gpt-5.2"
	MINIMAX_M2_5 OpenRouterModel = "minimax/minimax-m2.5"
)

func NewOpenRouterProvider() LLMProvider {
	return OpenRouterProvider{}
}

type OpenRouterRequest struct {
	Input       []OpenRouterRequestInput `json:"input"`
	Model       OpenRouterModel          `json:"model"`
	Temperature float64                  `json:"temperature"`
	TopP        float64                  `json:"top_p"`
}

type OpenRouterRequestInput struct {
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content string `json:"content"`
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
	Logprobs           any     `json:"logprobs"`
	FinishReason       string  `json:"finish_reason"`
	NativeFinishReason string  `json:"native_finish_reason"`
	Index              int     `json:"index"`
	Message            Message `json:"message"`
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

func (p OpenRouterProvider) Complete(systemPrompt string, userPrompt ...string) (string, error) {
	fmt.Println("Running OpenRouter api call")

	err := godotenv.Load()

	if err != nil {
		fmt.Println("Failed to load env file")
	}

	prompts := []OpenRouterRequestInput{{Content: systemPrompt, Role: "system"}}

	for _, prompt := range userPrompt {
		prompts = append(prompts, OpenRouterRequestInput{Content: prompt, Role: "user"})
	}

	// requestBody := OpenRouterRequest{
	// 	Model:       MINIMAX_M2_5,
	// 	Input:       prompts,
	// 	TopP:        0,
	// 	Temperature: 0,
	// }

	// value, err := json.Marshal(requestBody)

	testJsonString := `{
  "model": "openai/gpt-5.1-codex-mini:nitro",

  "messages": [
  {
      "role": "user",
      "content": "Write a function in typescript to find the median in a stream of data"
    }
  ]
}`

	req, err := http.NewRequest(http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader([]byte(testJsonString)))

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
	responseBody, err := io.ReadAll(res.Body)

	var openRouterOutputMap OpenRouterResponse

	err = json.Unmarshal(responseBody, &openRouterOutputMap)

	if err != nil {
		return "", err
	}

	fmt.Println(openRouterOutputMap.Choices[0].Message.Content)

	return openRouterOutputMap.Choices[0].Message.Content, nil
}
