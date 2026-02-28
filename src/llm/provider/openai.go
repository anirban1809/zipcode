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

	body := []byte(fmt.Sprintf(`
									{
        "model": "gpt-5-mini",
        "input": [
            {
                "role": "user",
                "content": "Create a simple 3 step plan on how to create a typescript function"
            }
        ]
    }`))

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/responses", bytes.NewReader(body))

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
