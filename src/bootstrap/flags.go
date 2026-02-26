package bootstrap

import llm "zipcode/src/llm/provider"

type StartupIntent struct {
	Workspace   string
	Prompt      string
	LLMProvider llm.LLMProvider
	Config      string
	Debug       bool
}

func (intent StartupIntent) ParseFlags() StartupIntent {
	return StartupIntent{}
}
