package llm

type Registry struct {
	Providers map[ProviderName]Provider
}

func NewRegistry() Registry {
	providers := map[ProviderName]Provider{
		OpenAIProvider:        &OpenAI{},
		OpenRouterAPIProvider: &OpenRouterProvider{},
		AnthropicProvider:     &Anthropic{},
	}

	return Registry{
		Providers: providers,
	}
}

func (r Registry) GetProvider(name ProviderName) Provider {
	return r.Providers[name]
}

func (r Registry) ProviderList() []ProviderName {
	return []ProviderName{OpenAIProvider, OpenRouterAPIProvider, AnthropicProvider}
}
