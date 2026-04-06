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
	GLM_5_V_TURBO          OpenRouterModel = "z-ai/glm-5v-turbo"
)

var ModelNames = []string{
	"openai/gpt-5.2",
	"minimax/minimax-m2.5",
	"minimax/minimax-m2.7",
	"anthropic/claude-sonnet-4.6",
	"anthropic/claude-haiku-4.5",
	"openai/gpt-5.1-codex-mini",
	"moonshotai/kimi-k2.5",
	"meta-llama/llama-3.3-70b-instruct",
	"z-ai/glm-4.7",
	"qwen/qwen3-coder-flash",
	"openai/gpt-5-nano",
	"z-ai/glm-5",
	"openai/gpt-5.4-nano",
	"deepseek/deepseek-v3.2",
	"openai/gpt-5.4",
	"openai/gpt-5.3-codex",
	"z-ai/glm-5v-turbo",
}

var ModelDescriptions = []string{
	"General-purpose GPT-5.2 model suitable for complex reasoning, coding, and multi-step tasks.",
	"Minimax M2.5 model optimized for fast inference and cost-efficient general tasks.",
	"Improved Minimax M2.7 model with better reasoning and language capabilities over M2.5.",
	"Claude Sonnet 4.6 model optimized for high-quality reasoning, coding, and structured outputs.",
	"Claude Haiku 4.5 model designed for fast, low-cost responses with good general performance.",
	"Lightweight GPT-5.1 Codex variant optimized for code generation and quick developer workflows.",
	"Kimi K2.5 model focused on long-context understanding and conversational reasoning.",
	"LLaMA 3.3 70B instruct-tuned model for general-purpose reasoning and instruction following.",
	"GLM 4.7 model offering balanced performance for multilingual and general AI tasks.",
	"Qwen3 Coder Flash model optimized for fast and efficient code generation tasks.",
	"Ultra-lightweight GPT-5 Nano model designed for very fast and low-cost inference.",
	"GLM-5 model with enhanced reasoning, multimodal capabilities, and improved accuracy.",
	"Improved nano-tier GPT-5.4 model with better efficiency and slightly enhanced reasoning.",
	"DeepSeek v3.2 model optimized for coding, reasoning, and cost-efficient inference.",
	"Latest GPT-5.4 model offering top-tier reasoning, coding, and multimodal capabilities.",
	"GPT-5.3 Codex model specialized for advanced code generation and developer workflows.",
	"GLM-5V Turbo model optimized for fast multimodal tasks including vision and text processing.",
}

func SetCurrentModel(model string) {
	CurrentModel = model
}

var CurrentModel = "openai/gpt-5.4"

var INTERNAL_TOOL_PATH = "/Users/anirban/Documents/Code/zipcode/src/tools"
var EXTERNAL_TOOL_PATH = "~/.zipcode/tools"
