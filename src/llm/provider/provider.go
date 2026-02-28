package llm

type LLMProvider interface {
	Complete(systemPrompt string, userPrompt string) (string, error)
}
