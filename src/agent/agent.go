package agent

import (
	llm "zipcode/src/llm/provider"
	"zipcode/src/tools"
)

type Agent struct {
	Name         string
	SystemPrompt string
	Conversation llm.Conversation
	LastResponse llm.Message
	Tools        *[]tools.Tool
	LLM          *llm.OpenRouterProvider
	initial      bool
}

func NewAgent(systemPrompt string, tools *[]tools.Tool, llm *llm.OpenRouterProvider) Agent {
	return Agent{
		SystemPrompt: systemPrompt,
		Tools:        tools,
		LLM:          llm,
		initial:      false,
	}
}

func (a *Agent) RunStep(messages ...llm.Message) (*llm.Conversation, error) {
	if !a.initial {
		a.Conversation = llm.Conversation{
			Messages: []llm.Message{
				{
					Content: a.SystemPrompt,
					Role:    "system",
				},
			},
			Tools: *a.Tools,
		}

		a.initial = true
	}

	a.Conversation.Messages = append(a.Conversation.Messages, messages...)
	a.Conversation.Tools = *a.Tools
	conv, err := a.LLM.Chat(&a.Conversation)

	if err != nil {
		return nil, err
	}

	a.Conversation = *conv
	return conv, nil
}
