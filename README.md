# ZipCode

A terminal-based AI assistant interface built with Go and the Bubble Tea framework. ZipCode provides a TUI (Terminal User Interface) for interacting with AI models through a clean, intuitive command-line interface.

## Overview

ZipCode is a single-binary Go application that simulates an AI assistant interface with:
- Interactive terminal UI with command suggestions
- Multiple AI model support (simulated)
- Workspace management
- Command system with autocomplete
- Real-time status display
- Responsive layout that adapts to terminal size

## Features

- **Terminal-first design**: Built with Bubble Tea for a smooth TUI experience
- **Command system**: Slash commands with autocomplete and descriptions
- **Model switching**: Switch between different AI models (Claude, Gemini, GPT)
- **Workspace awareness**: Displays current working directory
- **Interactive menus**: Navigate commands and models with keyboard
- **Mock responses**: Simulated AI responses with spinner animation
- **Usage tracking**: Displays session statistics (tokens, latency, uptime)

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

### Commands

Type `/` to see available commands. Use arrow keys to navigate and Enter to execute:

| Command | Description |
|---------|-------------|
| `/init` | Initialize the current workspace session |
| `/models` | Open model picker and switch active model |
| `/quit` | Exit the application |
| `/help` | Show help and available command usage |
| `/status` | Display current session and environment status |
| `/config` | View or adjust mock configuration settings |
| `/auth` | Manage mock authentication state |
| `/sync` | Run a mock workspace sync operation |
| `/deploy` | Start a mock deploy flow |
| `/logs` | Show recent mock execution logs |
| `/doctor` | Run mock diagnostics checks |
| `/clear` | Clear the result/output area |
| `/version` | Show application version details |
| `/theme` | Switch UI theme options (mock) |

### Available Models

- claude-4-sonnet
- claude-4-opus  
- gemini-2.0-flash
- gpt-4.1

## Architecture

### Project Structure

```
zipcode/
├── main.go              # Application entry point and TUI logic
├── go.mod              # Go module definition
├── Makefile            # Build and development targets
├── bin/                # Compiled binaries
├── dist/               # Release builds
└── src/                # Source code packages
    ├── ui/             # UI components and models
    ├── tools/          # Utility tools and integrations
    └── bootstrap/      # Application initialization
```

### Key Components

#### Main Application (`main.go`)
- **RootModel**: Main application model managing layout and terminal size
- **model**: Core TUI model handling user input, commands, and display
- **Command System**: Slash command parsing and execution
- **Menu Navigation**: Interactive menus for commands and model selection

#### UI Framework
- Built with **Bubble Tea** for reactive TUI
- Uses **Lipgloss** for styling and layout
- Responsive design that adapts to terminal dimensions

#### State Management
- Workspace tracking
- Current model selection
- Command history and suggestions
- Session statistics

## Dependencies

### Core Dependencies
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - UI components (textinput, viewport)
- `github.com/charmbracelet/lipgloss` - Styling and layout
- `golang.org/x/term` - Terminal handling

### Development Dependencies
- `golangci-lint` - Go linting and static analysis

## Development

### Code Style

The project follows standard Go conventions:
- `go fmt` for code formatting
- `golangci-lint` for static analysis
- Clear separation of concerns between UI and business logic

### Adding New Commands

1. Add command to `commands` slice in `initialModel()`
2. Add description to `commandDesc` map
3. Handle command in `runCommand()` switch statement

### Customizing Models

Modify the `modelOptions` slice in `initialModel()` to add or remove AI models.

## Build Metadata

The Makefile injects build metadata into the binary:
- Version: Git tag or "dev"
- Commit: Short Git hash or "none"  
- Build time: UTC timestamp

This metadata is displayed in the application footer and can be used for debugging.

## License

[Add your license information here]

## Contributing

[Add contribution guidelines here]

## Roadmap

- [ ] Real AI model integration
- [ ] Configuration file support
- [ ] Plugin system for extensions
- [ ] Persistent history
- [ ] Theme customization
- [ ] Multi-language support