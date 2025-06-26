# TAI (Terminal AI)

[![Go Report Card](https://goreportcard.com/badge/github.com/adamveld12/tai)](https://goreportcard.com/report/github.com/adamveld12/tai)
[![Go Reference](https://pkg.go.dev/badge/github.com/adamveld12/tai.svg)](https://pkg.go.dev/github.com/adamveld12/tai)
[![CI](https://github.com/adamveld12/tai/workflows/CI/badge.svg)](https://github.com/adamveld12/tai/actions)
[![codecov](https://codecov.io/gh/adamveld12/tai/branch/main/graph/badge.svg)](https://codecov.io/gh/adamveld12/tai)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/release/adamveld12/tai.svg)](https://github.com/adamveld12/tai/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/adamveld12/tai)](https://golang.org/)

A terminal-based AI assistant that provides both interactive REPL and one-shot modes for interacting with LLM providers. Built in Go using the Bubble Tea TUI framework.

## Features

- **REPL Mode**: Interactive terminal interface with conversation history
- **One-shot Mode**: Single command execution, perfect for scripting
- **Multiple LLM Providers**: Currently supports LMStudio (OpenAI-compatible)
- **Clean Architecture**: Redux-like state management with provider pattern
- **Thread-safe**: Concurrent operations with proper synchronization

## Quick Start

### Install from Binary

Download the latest binary from the [releases page](https://github.com/adamveld12/tai/releases) and place it in your PATH.

### Install from Source

```bash
git clone https://github.com/adamveld12/tai.git
cd tai
make build
sudo mv build/tai /usr/local/bin/
```

### Running

**REPL Mode (Interactive):**

```bash
tai
```

**One-shot Mode:**

```bash
tai --oneshot "What's the weather like?"
echo "Explain this code" | tai --oneshot
tai --oneshot "Summarize this:" < file.txt
```

## Development Setup

### Prerequisites

- Go 1.24.4 or later
- LMStudio running on `localhost:1234` (default provider)

### Getting Started

```bash
# Clone the repository
git clone https://github.com/adamveld12/tai.git
cd tai

# Install dependencies
make deps

# Build the project
make build

# Run the application
make run

# Run tests
make test

# Run with race detection
make test-race

# Check code quality
make check
```

### Development Commands

```bash
make build              # Build binary to ./build/tai
make run                # Run the application
make test               # Run tests with coverage
make test-race          # Run tests with race detection
make check              # Run all quality checks (fmt, vet, lint, test)
make clean              # Clean build artifacts
make install            # Install to $GOPATH/bin
```

### LMStudio Setup

1. Download and install [LMStudio](https://lmstudio.ai/)
2. Load your preferred model
3. Start the local server (default: `http://localhost:1234/v1`)
4. Run TAI - it will automatically connect

Alternative: Use the included helper:

```bash
make lmstudio          # Start LMStudio server
```

## Architecture

TAI follows a layered architecture with Redux-like state management:

```
cmd/tai/main.go    → Entry point and mode selection
internal/cli/      → Configuration and one-shot handler
internal/ui/       → Bubble Tea UI components (REPL)
internal/state/    → Redux-like state management
internal/llm/      → Provider interface and implementations
```

### Key Components

- **State Management**: Immutable state updates with thread-safe dispatching
- **Provider Pattern**: Pluggable LLM providers implementing a common interface
- **Mode Separation**: REPL for interactive use, one-shot for scripting
- **UI Components**: Modular Bubble Tea components with clean separation

## Configuration

TAI uses sensible defaults but can be configured:

- **LLM Provider**: Currently LMStudio at `http://localhost:1234/v1`
- **Models**: Automatically detects available models from provider
- **REPL Commands**: `:help`, `:clear`, `:quit`

## Testing

The project emphasizes production confidence with comprehensive testing:

```bash
make test              # Run all tests with coverage
make test-race         # Run with race detection
make test-coverage     # Generate HTML coverage report
```

Current coverage: 92.7% (LLM providers), 80% (state management)

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make changes and add tests
4. Run quality checks: `make check`
5. Submit a pull request

## Roadmap

- [ ] Additional LLM providers (OpenAI, Anthropic, Ollama)
- [ ] Tool system for file operations and shell execution
- [ ] Enhanced logging and formatting
- [ ] Configuration file support
- [ ] CI/CD and automated releases

## License

MIT License - see LICENSE.md file for details.

## Related Projects

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [LMStudio](https://lmstudio.ai/) - Local LLM runtime
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
