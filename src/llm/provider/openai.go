package llm

type OpenAI struct {
	ProviderId string
	Model      string
}

func (model OpenAI) ID() string {
	return "openai"
}
