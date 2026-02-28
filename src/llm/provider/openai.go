package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"zipcode/src/llm/prompts"

	"github.com/joho/godotenv"
)

type OpenAI struct {
	ProviderId string
	Model      string
}

func NewOpenAIProvider() LLMProvider {
	return OpenAI{
		ProviderId: "openai",
		Model:      "gpt-5-mini",
	}
}

type Input struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model string  `json:"model"`
	Input []Input `json:"input"`
}

type Response struct {
	Output []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func (model OpenAI) Complete(systemPrompt string, userPrompt string) (string, error) {
	fmt.Println("Running open ai call...")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Failed to load env file")
	}

	requestBody := Request{
		Model: "gpt-5-mini",
		Input: []Input{{
			Content: prompts.IntentClassifier,
			Role:    "system",
		}, {
			Content: "Login fails when password contains special characters",
			Role:    "user",
		}},
	}

	// requestBody.Input[0].Content = "Create a simple 3 step plan on how to create a python function"
	// requestBody.Input[0].Role = "user"

	value, err := json.Marshal(requestBody)

	// body := fmt.Appendf(nil, `
	// 								{
	//     "model": "gpt-5-mini",
	//     "input": [
	//         {
	//             "role": "user",
	//             "content": "Create a simple 3 step plan on how to create a typescript function"
	//         }
	//     ]
	// }`)

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/responses", bytes.NewReader(value))

	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	responseBody, err := io.ReadAll(res.Body)

	var outputMap Response

	err = json.Unmarshal(responseBody, &outputMap)

	if err != nil {
		return "", err
	}

	return outputMap.Output[1].Content[0].Text, nil

}
