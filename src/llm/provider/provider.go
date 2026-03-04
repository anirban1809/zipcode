package llm

type LLMProvider interface {
	SetModel(model OpenRouterModel, nitro bool)
	Complete(systemPrompt string, userPrompt ...string) (string, error)
}
