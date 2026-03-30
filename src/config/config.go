package config

var HEADLESS = false
var APP_VERSION = "0.0.1"

type OpenRouterModel string

const (
	GPT_5_2                OpenRouterModel = "openai/gpt-5.2"
	MINIMAX_M2_5           OpenRouterModel = "minimax/minimax-m2.5"
	MINIMAX_M2_7           OpenRouterModel = "minimax/minimax-m2.7"
	CLAUDE_SONNET_4_6      OpenRouterModel = "anthropic/claude-sonnet-4.6"
	CLAUDE_HAIKU_4_5       OpenRouterModel = "anthropic/claude-haiku-4.5"
	GPT_5_1_CODEX_MINI     OpenRouterModel = "openai/gpt-5.1-codex-mini"
	KIMI_K_2_5             OpenRouterModel = "moonshotai/kimi-k2.5"
	LLAMA_3_3_70B_INSTRUCT OpenRouterModel = "meta-llama/llama-3.3-70b-instruct"
	GLM_4_7                OpenRouterModel = "z-ai/glm-4.7"
	QWEN_3_CODER_FLASH     OpenRouterModel = "qwen/qwen3-coder-flash"
	GPT_5_NANO             OpenRouterModel = "openai/gpt-5-nano"
	GLM_5                  OpenRouterModel = "z-ai/glm-5"
	GPT_5_4_NANO           OpenRouterModel = "openai/gpt-5.4-nano"
	DEEPSEEK_3_2           OpenRouterModel = "deepseek/deepseek-v3.2"
	GPT_5_4                OpenRouterModel = "openai/gpt-5.4"
	GPT_5_3_CODEX          OpenRouterModel = "openai/gpt-5.3-codex"
)

var CurrentModel = GPT_5_3_CODEX
