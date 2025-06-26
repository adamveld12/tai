# TAI Architecture

## Overview

TAI (Terminal AI) is a terminal-based AI assistant built with Go, using the Bubble Tea framework for the TUI and a Redux-like state management pattern.

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                          CLI Layer                          │
│                      (cmd/tai/main.go)                     │
├─────────────────────────────────────────────────────────────┤
│                      UI Layer                               │
│                   (internal/ui/)                           │
│   ┌─────────────────┬─────────────────┬─────────────────┐   │
│   │   Header View   │  Message View   │   Input View    │   │
│   └─────────────────┴─────────────────┴─────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│                   State Management                          │
│                   (internal/state/)                        │
│   ┌─────────────────┬─────────────────┬─────────────────┐   │
│   │     Store       │    Actions      │    Reducers     │   │
│   └─────────────────┴─────────────────┴─────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│              Business Logic Layer                           │
│   ┌─────────────────┬─────────────────┬─────────────────┐   │
│   │  LLM Providers  │  Tool System    │  CLI Commands   │   │
│   │ (internal/llm/) │(internal/tools/)│(internal/cli/)  │   │
│   └─────────────────┴─────────────────┴─────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│                    Models & Utils                           │
│   ┌─────────────────┬─────────────────────────────────────┐ │
│   │  Data Models    │         Utilities                   │ │
│   │ (pkg/models/)   │       (pkg/utils/)                  │ │
│   └─────────────────┴─────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Component Details

### CLI Layer (`cmd/tai/`)
- Entry point for the application
- Handles command-line argument parsing
- Initializes the application and starts the appropriate mode (REPL or one-shot)

### UI Layer (`internal/ui/`)
- Built with Bubble Tea framework
- Implements the Model-View-Update pattern
- Components:
  - **Header**: Shows current provider, model, and status
  - **Message View**: Displays conversation history with syntax highlighting
  - **Input View**: Handles user input with autocompletion and command suggestions
  - **Sidebar**: Optional panel for history, settings, and help

### State Management (`internal/state/`)
- Redux-like pattern with centralized state
- **Store**: Holds the application state
- **Actions**: Define state mutations (SendMessage, ReceiveMessage, etc.)
- **Reducers**: Pure functions that update state based on actions
- **Middleware**: Handles side effects like API calls and persistence

### LLM Providers (`internal/llm/`)
- Abstracted provider interface
- Implementations for:
  - OpenAI GPT models
  - Google Gemini
  - Ollama (local models)
  - LMStudio (local inference server)
- Streaming and non-streaming chat completions
- Model listing and validation

### Tool System (`internal/tools/`)
- Plugin-like architecture for tools
- Core tools:
  - **File Tool**: Read, write, list, delete files
  - **Shell Tool**: Execute shell commands
  - **Search Tool**: Search files and content
  - **Network Tool**: HTTP requests, downloads
- Tool registry for dynamic tool management

### CLI Commands (`internal/cli/`)
- Colon commands for REPL mode (`:help`, `:quit`, `:clear`, etc.)
- Command parsing and validation
- Autocompletion support

### Models (`pkg/models/`)
- Data structures for configuration, conversations, sessions
- JSON/YAML serialization support
- Type safety for all data exchanges

### Utilities (`pkg/utils/`)
- Common utility functions
- Configuration loading/saving
- Logging utilities
- File system helpers

## Data Flow

### REPL Mode Flow
```
User Input → UI Layer → State (Action) → Reducer → LLM Provider → State Update → UI Render
```

### One-shot Mode Flow
```
CLI Args → Parse → LLM Provider → Format Output → Exit
```

### Tool Execution Flow
```
LLM Response → Tool Parser → Tool Registry → Tool Execution → Result → State Update
```

## State Structure

```go
type AppState struct {
    // UI state
    UI UIState
    
    // Chat state
    Chat ChatState
    
    // Configuration
    Config AppConfig
    
    // Session information
    Session Session
    
    // Tool registry
    Tools ToolRegistry
    
    // Provider registry
    Providers map[string]Provider
}
```

## Key Design Patterns

### 1. Repository Pattern
- Abstract data access for conversations, history, configuration
- Pluggable backends (file system, database, etc.)

### 2. Strategy Pattern
- LLM providers implement common interface
- Tools implement common interface
- Output formatters implement common interface

### 3. Observer Pattern
- UI components observe state changes
- Automatic re-rendering on state updates

### 4. Command Pattern
- Colon commands encapsulate actions
- Undo/redo capability for certain operations

### 5. Middleware Pattern
- State middleware for logging, persistence, validation
- HTTP middleware for provider requests

## Configuration

Configuration is loaded from multiple sources in order of precedence:
1. Command-line flags
2. Environment variables
3. Configuration file (`~/.tai/config.yaml`)
4. Default values

## Security Considerations

- API keys stored securely with optional encryption
- Tool execution sandboxing
- Input validation and sanitization
- Rate limiting for LLM providers
- Audit logging for tool executions

## Performance Optimizations

- Lazy loading of providers and tools
- Message streaming for better responsiveness
- Efficient text rendering with virtual scrolling
- Connection pooling for HTTP requests
- Caching of provider responses (optional)

## Testing Strategy

- Unit tests for all pure functions (reducers, utilities)
- Integration tests for provider implementations
- UI tests using Bubble Tea testing utilities
- End-to-end tests for CLI workflows
- Benchmark tests for performance-critical paths
