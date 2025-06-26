# TAI (Terminal AI) Development Plan

## Overview
TAI is a terminal-based AI assistant that provides both REPL (interactive) and one-shot modes for interacting with various LLM providers.

## Development Phases

### 1. Project Initialization
- Initialize the Go module in `/Users/adam/code/tai` with `go mod init github.com/adamveld12/tai`
- Create core directories: `cmd/tai`, `internal/{cli,ui,llm,tools,state}`, `pkg/{models,utils}`
- Add a `Makefile` with tasks for build, test, and lint

### 2. Define Core Architecture & Interfaces
- Draft architecture diagrams outlining layers (CLI, UI, state, LLM, tools)
- Define Go interfaces for LLM providers (e.g., `Provider`, `ChatCompletion`)
- Define interfaces for tool actions (e.g., `FileTool`, `ShellTool`)

### 3. Setup State Management (Redux Pattern)
- Implement a central store with state, actions, and reducers in `internal/state`
- Define action types (`SendMessage`, `ReceiveMessage`, `Clear`, etc.)
- Write pure reducers to update store immutably

### 4. Implement CLI Framework with Bubble Tea
- Scaffold the Bubble Tea application in `cmd/tai/main.go`
- Create UI components: header, message view, input box
- Wire the Tea `Model`, `Init`, `Update`, and `View` methods

### 5. Develop REPL Mode
- Implement multi-turn chat loop in UI with color-coded roles
- Support colon commands (`:help`, `:quit`, `:clear`) in the input handler
- Add autocompletion and command suggestions

### 6. Implement One-shot Mode
- Parse `--oneshot` flag and read input from args or stdin
- Execute single prompt and exit with structured JSON output
- Ensure exit codes reflect success or error

### 7. Integrate LLM Providers
- Create provider implementations for OpenAI, Google Gemini, Ollama, LMStudio
- Abstract common logic behind the `Provider` interface
- Add config loading and credentials support

### 8. Build Tool System
- Develop file operations: `List`, `Read`, `Write`, `Delete`, `Rename`
- Implement shell execution tools using `os/exec`
- Support recursive sub-agent calls and research/search tool

### 9. Enhance Logging, Error Handling & Formatting
- Integrate timestamped logs for all messages
- Use `lipgloss` or `glamour` for rich formatting and colors
- Display structured errors in the UI

### 10. Testing & Code Coverage
- Write unit tests for reducers, providers, tools, and UI logic
- Aim for at least 75% coverage using `go test -coverprofile`
- Configure GitHub Actions for automated test runs

### 11. Documentation & Examples
- Create `README.md` with installation, usage, and examples for REPL and one-shot
- Add code comments and GoDoc for public APIs
- Provide sample config and conversation transcripts

### 12. CI/CD & Release Setup
- Configure GitHub Actions workflows for linting (`golangci-lint`), testing, and releasing
- Define semantic versioning and draft GitHub release templates
- Publish binaries for major platforms

## Current Status
- [x] Phase 1: Project Initialization
- [x] Phase 2: Define Core Architecture & Interfaces
- [x] Phase 3: Setup State Management
- [x] Phase 4: Implement CLI Framework
- [x] Phase 5: Develop REPL Mode
- [ ] Phase 6: Implement One-shot Mode
- [ ] Phase 7: Integrate LLM Providers
- [ ] Phase 8: Build Tool System
- [ ] Phase 9: Enhance Logging & Formatting
- [ ] Phase 10: Testing & Code Coverage
- [ ] Phase 11: Documentation & Examples
- [ ] Phase 12: CI/CD & Release Setup
