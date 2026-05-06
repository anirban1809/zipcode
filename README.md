# ZipCode

A terminal-based AI coding assistant built with Go and tuix. ZipCode provides a TUI (Terminal User Interface) for interacting with AI models through OpenRouter to perform code exploration, file manipulation, and task execution.

## Overview

ZipCode is a single-binary Go application that provides an AI coding assistant interface with:
- Interactive terminal UI built with tuix
- Real AI model integration via OpenRouter
- Tool execution (file operations, shell commands, code search)
- Sub-agent support for specialized tasks
- Workspace awareness and management

## Features

- **Terminal-first design**: Built with tuix for a smooth TUI experience
- **Real AI integration**: Connects to OpenRouter for LLM interactions
- **Tool system**: Execute file operations, shell commands, and code searches
- **Sub-agents**: Specialized agents for code exploration and bug investigation
- **Multiple model support**: Switch between various AI models via OpenRouter
- **Workspace management**: Track and manage the current working directory
- **Interactive UI**: Command input, file diffs, status display

## Installation

### Prerequisites

- Go 1.24.2 or later
- Git (for build metadata)

### Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd zipcode

# Build the binary
make build

# Run the application
./bin/zipcode
```

### Development

```bash
# Run in development mode with race detector
make dev

# Run tests
make test

# Format code
make fmt

# Run linter
make lint

# Clean build artifacts
make clean

# Build release binaries for multiple platforms
make dist
```

## Usage

### Starting the Application

```bash
# Run from source
make run

# Or run the built binary
./bin/zipcode
```

### Basic Controls

- **Enter**: Submit prompt or execute selected command
- **Shift+Enter**: Add newline to prompt
- **↑/↓ Arrow Keys**: Navigate menus or scroll results
- **Tab**: Select highlighted menu item
- **Ctrl+C / q / Esc**: Quit application

### Available Models

ZipCode connects to OpenRouter and supports 17+ models including:
- minimax/minimax-m2.5 (default)
- openai/gpt-5.1
- anthropic/claude-4.5
- google/gemini-2.5-pro
- deepseek/deepseek-chat
- meta-llama/llama-4
- And many more...

### Tools

The assistant can use various tools to help with tasks:

| Tool | Description |
|------|-------------|
| `file_read` | Read file contents from the workspace |
| `file_write` | Write or create files in the workspace |
| `bash` | Execute shell commands |
| `codesearch` | Search for code patterns in files |
| `filesearch` | Find files by name pattern |

### Sub-agents

Specialized agents for complex tasks:

- **code_explorer**: Analyze and understand codebase structure
- **bug_investigator**: Identify bugs and suggest fixes

## Architecture

### Project Structure

```
zipcode/
├── main.go              # Application entry point
├── go.mod               # Go module definition
├── Makefile             # Build and development targets
├── bin/                 # Compiled binaries
└── src/
    ├── agent/           # Core agent runtime and executor
    ├── bootstrap/       # Application initialization and flags
    ├── config/          # Configuration (models, paths)
    ├── llm/             # LLM providers and prompts
    ├── subagents/       # Sub-agent definitions
    ├── tools/           # Tool implementations
    ├── ui/              # Deprecated bubbletea UI
    ├── utils/           # Utility functions
    ├── view/            # tuix UI components
    └── workspace/       # Workspace management
```

### Key Components

#### Main Application (`main.go`)
- Initializes terminal size
- Creates workspace from current directory
- Initializes agent runtime
- Launches tuix-based TUI application

#### Agent Runtime (`src/agent/`)
- **Runtime**: Manages the tool-calling loop with LLM
- **Executor**: Processes and executes tool calls
- **Planner**: Coordinates task execution

#### LLM Integration (`src/llm/`)
- OpenRouter provider integration
- Prompt management
- Model configuration

#### UI Framework (`src/view/`)
- Built with **tuix** for reactive TUI
- Uses **Lipgloss** for styling
- Components: app, menu, prompt, statusline, filediff

## Dependencies

### Core Dependencies
- `github.com/anirban1809/tuix` - TUI framework
- `golang.org/x/term` - Terminal handling

### Development Dependencies
- `golangci-lint` - Go linting and static analysis

## Configuration

Environment variables (see `.env`):
- `OPENROUTER_API_KEY` - API key for OpenRouter
- `OPENROUTER_BASE_URL` - Optional custom endpoint
- Internal/external tool paths configured in `src/config/`

## Build Metadata

The Makefile injects build metadata into the binary:
- Version: Git tag or "dev"
- Commit: Short Git hash or "none"
- Build time: UTC timestamp

## License

[Add your license information here]

## Contributing

[Add contribution guidelines here]