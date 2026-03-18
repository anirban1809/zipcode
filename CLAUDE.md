# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build    # Build binary to ./bin/zipcode
make run      # Run with go run
make dev      # Run with race detector
make test     # Run all tests (go test ./...)
make lint     # Run golangci-lint
make fmt      # Format code
make clean    # Remove bin/ and dist/
make dist     # Build release binaries for Linux/macOS
```

## Architecture

ZipCode is a terminal AI coding assistant (TUI) written in Go. It uses [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the UI and connects to LLM providers via OpenRouter to run tool-using agents.

### Data Flow

```
User Input → AppModel (UI) → Runtime (agent loop) → Executor (tool dispatch)
  → LLM (OpenRouter) → Tools (bash/file/search) → ResponseEvent → UI update
```

### Key Components

**`src/ui/appmodel.go`** — Central TUI state machine. Handles key events, manages task list, renders viewport. On prompt submit: creates a Task, starts a goroutine calling `Runtime.Run()`, then listens on `GetExecutorEventChannel()` for `ResponseEvent` messages. Two event types: `Tool` (tool call in progress) and `Message` (final response).

**`src/agent/runtime.go`** — Agent loop. Builds conversation from system prompt + user request, calls `LLM.Chat()`, then loops: `Executor.ProcessResponse()` → append tool result → `LLM.Chat()` → repeat until complete.

**`src/agent/executor.go`** — Tool dispatch. Routes tool calls to implementations, pushes `ResponseEvent` to UI via `EventChannel`. Can push interactive questions and wait on `MessageChannel` for user responses.

**`src/tools/`** — Tool implementations: `bash_tool`, `file_read`, `file_search` (uses `rg`), `file_write`, `code_search`. All follow the OpenAI function-calling schema in `common.go`.

**`src/llm/provider/openrouter.go`** — OpenRouter API client. Supported models include Claude Sonnet 4.6, various GPT/Llama/Qwen models. Default is Minimax M2.5.

**`src/workspace/`** — Workspace context: root path, config, history, metadata. Currently hardcoded path in `main.go`.

**`src/ui/components/`** — `task.go` (spinner + result display), `statusbar.go` (IDLE/RUNNING/ERROR + workspace/model info), `ask.go` (interactive option selection for agent questions).

### Runtime Requirements

- Go 1.24.2+
- `rg` (ripgrep) for file search tool
- OpenRouter API key in environment
