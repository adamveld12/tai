# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TAI (Terminal AI) is a terminal-based AI assistant that provides both REPL (interactive) and one-shot modes for interacting with various LLM providers. It's built in Go using the Bubble Tea TUI framework.

**Key Dependencies:**
- `github.com/charmbracelet/bubbletea` - TUI framework for REPL mode
- `github.com/sashabaranov/go-openai` - OpenAI API client (used by LMStudio provider)
- `github.com/stretchr/testify` - Testing framework with assertions and mocking

## Development Commands

### Build & Run

```bash
make build              # Build the binary to ./build/tai
make run                # Run the application
make dev-reload         # Run with live reload using air (if installed)
make lmstudio          # Start LMStudio server (default LLM provider)
```

### Testing

```bash
make test              # Run all tests with coverage (includes -v flag)
make test-race         # Run tests with race detection (includes -v flag)
make test-coverage     # Run tests with coverage report and generate HTML (includes -v flag)
make test-watch        # Watch files and re-run tests on changes (requires air)
go test ./internal/cli -v -run TestOneShotHandler  # Run specific test
go test ./internal/state -v  # Run state package tests
go test ./internal/llm -v   # Run LLM provider tests (comprehensive suite)
go test ./internal/llm -v -run TestChatCompletion_SuccessScenarios  # Run specific test group
```

### Code Quality

```bash
make check             # Run all quality checks (fmt, vet, lint, test, test-race)
make fmt               # Format code
make lint              # Run linter (requires golangci-lint)
make vet               # Run go vet
```

### Installation & Cleanup

```bash
make install           # Install tai to $GOPATH/bin
make build-all         # Build for multiple platforms
make clean             # Clean build artifacts and coverage files
make deps              # Download dependencies
```

## Architecture Overview

TAI follows a layered architecture with Redux-like state management:

```
cmd/tai/main.go → Mode Selection (REPL or One-shot)
                      ↓
internal/cli/    → Configuration and one-shot handler
                      ↓
internal/ui/     → Bubble Tea UI components (REPL mode)
                      ↓
internal/state/  → Redux-like state management with Actions
                      ↓
internal/llm/    → Provider interface and implementations
```

### Key Architectural Patterns

1. **State Management**: Central `AppState` with immutable updates via Actions and Dispatchers. Thread-safe implementation using mutex locks.

2. **Provider Pattern**: All LLM providers implement the `Provider` interface:
   - `ChatCompletion()` for synchronous requests
   - `StreamChatCompletion()` for streaming responses
   - Currently only `LMStudioProvider` is implemented

3. **Mode Separation**:
   - **REPL Mode**: Full TUI with conversation history, uses state management
   - **One-shot Mode**: Direct LLM call, no state persistence, suitable for scripting

### Working with State

State updates follow the Redux pattern:

1. Create an Action that implements `state.Action` interface
2. Implement `Execute(AppState) (AppState, error)` method
3. Dispatch via `dispatcher.Dispatch(action)`

Example action pattern (note: no actions.go file exists yet):

```go
type SendMessageAction struct {
    Content string
}

func (a *SendMessageAction) Execute(state AppState) (AppState, error) {
    newState := state // Copy state
    newState.Context.Messages = append(newState.Context.Messages, Message{
        Role:    RoleUser,
        Content: a.Content,
    })
    return newState, nil
}
```

**State Management Architecture:**
- `MemoryState` implements the `Dispatcher` interface
- All state updates are thread-safe using `sync.RWMutex`
- Listeners are notified asynchronously via goroutines
- State immutability is enforced - `GetState()` returns copies, not references

### Adding New LLM Providers

1. Implement the `llm.Provider` interface in `internal/llm/`
2. Add provider initialization in `NewOneShotHandler()` and UI initialization
3. Provider interface requires:
   - `Name() string`
   - `ChatCompletion(ctx, req) (*ChatResponse, error)`
   - `StreamChatCompletion(ctx, req) (<-chan ChatStreamChunk, error)`
   - `Models(ctx) ([]string, error)`

Example provider implementation pattern:
```go
type MyProvider struct {
    client *http.Client
    config ProviderConfig
}

func (p *MyProvider) ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
    // Implementation with proper error handling and retries
}
```

### UI Components

The REPL UI uses Bubble Tea's Model-View-Update pattern:

- `internal/ui/app.go`: Main application model
- `internal/ui/header.go`: Header component
- `internal/ui/messages.go`: Conversation display
- `internal/ui/input.go`: User input handling

Colon commands are handled in the input component (`:help`, `:clear`, `:quit`).

### Testing Approach

- **LLM Provider Tests**: Comprehensive HTTP-level mocking with real request/response simulation
- **State Management Tests**: Redux pattern validation with thread-safety verification
- Mock implementations for `Provider` and `Dispatcher` interfaces using testify
- Table-driven tests with clear success/error scenario separation
- **Always run `make test-race`** to detect concurrency issues
- Current coverage: 92.7% (LLM provider), 80% (state management)
- Testing focuses on public API behavior, not implementation details

## Important Implementation Details

1. **One-shot Mode**: When both stdin and CLI prompt are provided, they're combined with the prompt first, then stdin content.

2. **State Persistence**: Currently uses in-memory state. Snapshot functionality is defined in interfaces but not yet implemented.

3. **Provider Configuration**: LMStudio provider expects server at `http://localhost:1234/v1`

4. **Thread Safety**: State updates use mutex locks. All state modifications must go through the Dispatcher.

5. **Error Handling**: Provider errors are wrapped with context. One-shot mode returns non-zero exit codes on failure.

6. **Retry Logic**: LMStudio provider retries all errors except specific strings (`invalid_api_key`, `model_not_found`) with exponential backoff.

7. **Streaming**: Proper goroutine cleanup and context cancellation handling in StreamChatCompletion.

## Common Troubleshooting

### LMStudio Connection Issues
- Ensure LMStudio is running: `make lmstudio`
- Check server is accessible: `curl http://localhost:1234/v1/models`
- Verify model is loaded in LMStudio UI

### Build Errors
- Run `make deps` to ensure all dependencies are downloaded
- Check Go version: requires Go 1.21+
- Clear build cache: `go clean -cache` if experiencing stale builds

### Test Failures
- Race conditions: Always use `make test-race` during development
- Flaky tests: Check for proper context cancellation and goroutine cleanup
- Mock server issues: Ensure no port conflicts on :1234

## Debugging Tips

### Using Delve with Bubble Tea
```bash
# Debug REPL mode
dlv debug ./cmd/tai -- -repl

# Debug one-shot mode
dlv debug ./cmd/tai -- "your prompt here"
```

### State Inspection
- Add debug logging in Action.Execute() methods
- Use `dispatcher.GetState()` to inspect current state
- Monitor state changes by implementing a debug Listener

### Performance Profiling
```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./internal/llm
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./internal/state
go tool pprof mem.prof
```

## Current Development Status

Based on PLAN.md:

- ✅ Phases 1-6 complete (Project setup through One-shot mode)
- ✅ Phase 10: Testing & Code Coverage (92.7% coverage for LLM, 80% for state)
- ⏳ Phases 7-9, 11-12: Provider integrations, Tool system, Logging, Documentation, CI/CD

## Key Files for Understanding the Codebase

- `PLAN.md` - Development roadmap and phase completion status
- `TESTING_PLAN.md` - Testing strategy overhaul (completed: comprehensive LLM provider tests)
- `ARCHITECTURE.md` - Detailed architecture documentation
- `internal/state/` - Redux-like state management with comprehensive tests
- `internal/llm/lmstudio_test.go` - Production-quality test suite with HTTP mocking and race detection
- `internal/cli/oneshot.go` - One-shot mode implementation with message construction
- `Makefile` - All development commands including race detection

## Testing Strategy

The test suite emphasizes **production confidence** over coverage metrics:

### LLM Provider Testing (92.7% coverage)
- **HTTP-level mocking** using `httptest.Server` for realistic testing  
- **Comprehensive scenarios**: Success, errors, retries, streaming, tool calls
- **Edge cases**: Context cancellation, goroutine leaks, large responses, Unicode
- **Actual retry behavior validation**: Tests discovered retry logic uses error strings, not HTTP codes
- **Race condition detection**: All tests pass with `-race` flag

### Writing New Provider Tests
```go
// Use proper test infrastructure
provider := newTestProvider(t, ProviderConfig{BaseURL: mock.URL()})

// Test real behavior, not implementation details  
resp, err := provider.ChatCompletion(ctx, request)

// Verify actual request/response flow
assert.Equal(t, expectedRequests, mock.RequestCount())
```

### Key Testing Principles Applied
- Test public API behavior, not internal methods
- Use realistic HTTP mocking, not interface mocks for network calls
- Verify actual timing for retry backoff and context cancellation
- Document WHY each test exists in clear comments
- Keep test organization clean with helper functions

## Code Style Guidelines

### Error Handling
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to complete chat: %w", err)
}
```

### Action Naming
- Use descriptive verb-noun pattern: `SendMessageAction`, `ClearHistoryAction`
- Keep actions focused on single state changes

### Public API Documentation
```go
// ChatCompletion sends a chat request to the LLM provider and returns the response.
// It handles retries with exponential backoff for transient errors.
// Context cancellation is properly handled to avoid goroutine leaks.
func (p *Provider) ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
```

## Performance Considerations

### Memory Management
- State copies are created on each update - consider pooling for high-frequency updates
- Message history grows unbounded - implement trimming for long conversations
- Stream chunks are sent through unbuffered channels - consider buffering for performance

### Concurrency
- State updates are serialized through mutex - consider read-write lock optimization
- Listeners are notified asynchronously - monitor goroutine growth
- HTTP clients are reused - ensure proper connection pooling

### Stream Processing
- Use context cancellation to stop streams early
- Implement backpressure handling for slow consumers
- Monitor memory usage with large streaming responses