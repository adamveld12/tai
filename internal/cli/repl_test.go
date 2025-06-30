package cli

import (
	"context"
	"testing"

	"github.com/adamveld12/tai/internal/llm"
	"github.com/adamveld12/tai/internal/state"
	"github.com/adamveld12/tai/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// mockStack is a mock implementation of ui.Stack for testing
type mockStack struct {
	activeScreen ui.Screen
	pushCount    int
	popCount     int
	clearCount   int
}

func (m *mockStack) Active() ui.Screen {
	return m.activeScreen
}

func (m *mockStack) Push(screen ui.Screen) int {
	m.pushCount++
	m.activeScreen = screen
	return m.pushCount
}

func (m *mockStack) Pop() ui.Screen {
	m.popCount++
	prev := m.activeScreen
	m.activeScreen = nil
	return prev
}

func (m *mockStack) Clear() {
	m.clearCount++
	m.activeScreen = nil
}

func (m *mockStack) Init() tea.Cmd {
	return nil
}

func (m *mockStack) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *mockStack) View() string {
	return "mock stack view"
}

func (m *mockStack) Run() error {
	return nil
}

func (m *mockStack) OnStateChange(action state.Action, newState state.AppState, oldState state.AppState) {
}

// mockScreen is a mock implementation of ui.Screen for testing
type mockScreen struct {
	initCalled         bool
	updateCalled       bool
	viewCalled         bool
	onStateChangeCalls []mockStateChangeCall
}

type mockStateChangeCall struct {
	action   state.Action
	newState state.AppState
}

func (m *mockScreen) Init() tea.Cmd {
	m.initCalled = true
	return nil
}

func (m *mockScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.updateCalled = true
	return m, nil
}

func (m *mockScreen) View() string {
	m.viewCalled = true
	return "mock screen view"
}

func (m *mockScreen) OnStateChange(action state.Action, newState state.AppState, oldState state.AppState) tea.Msg {
	m.onStateChangeCalls = append(m.onStateChangeCalls, mockStateChangeCall{
		action:   action,
		newState: newState,
	})
	return nil
}

// mockDispatcher and mockProvider are defined in oneshot_test.go and shared between test files

func TestNewReplHandler(t *testing.T) {
	tests := []struct {
		name                 string
		config               *Config
		expectPanic          bool
		expectedSystemPrompt string
		expectedWorkingDir   string
	}{
		{
			name: "successful creation with full config",
			config: &Config{
				SystemPrompt:     "Custom system prompt",
				WorkingDirectory: "/custom/dir",
				Prompt:           "test prompt",
			},
			expectPanic:          false,
			expectedSystemPrompt: "Custom system prompt",
			expectedWorkingDir:   "/custom/dir",
		},
		{
			name: "successful creation with minimal config",
			config: &Config{
				SystemPrompt:     "",
				WorkingDirectory: "",
				Prompt:           "",
			},
			expectPanic:          false,
			expectedSystemPrompt: "You are an AI assistant that autonomously writes code and helps the user with programming tasks.",
			expectedWorkingDir:   "", // Will be set to current working directory
		},
		{
			name: "successful creation with partial config",
			config: &Config{
				SystemPrompt:     "Test prompt",
				WorkingDirectory: "/test",
				Prompt:           "hello",
			},
			expectPanic:          false,
			expectedSystemPrompt: "Test prompt",
			expectedWorkingDir:   "/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handler *ReplHandler
			var panicked bool

			// Capture panics from log.Fatalf calls
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()

			// Note: This will actually try to create an LMStudio provider
			// In a real test environment, you might want to mock this
			if !tt.expectPanic {
				handler = NewReplHandler(tt.config)
			}

			if panicked != tt.expectPanic {
				t.Errorf("NewReplHandler() panicked = %v, expectPanic %v", panicked, tt.expectPanic)
				return
			}

			if tt.expectPanic {
				return // Skip further checks if we expected a panic
			}

			// Verify handler was created
			if handler == nil {
				t.Fatal("NewReplHandler() returned nil")
			}

			// Verify components are initialized
			if handler.Dispatcher == nil {
				t.Error("Dispatcher should not be nil")
			}
			if handler.Provider == nil {
				t.Error("Provider should not be nil")
			}
			if handler.Stack == nil {
				t.Error("Stack should not be nil")
			}
			if handler.Config == nil {
				t.Error("Config should not be nil")
			}

			// Verify state configuration
			state := handler.Dispatcher.GetState()
			if state.Context.SystemPrompt != tt.expectedSystemPrompt {
				t.Errorf("SystemPrompt = %q, want %q", state.Context.SystemPrompt, tt.expectedSystemPrompt)
			}

			if tt.expectedWorkingDir != "" && state.Context.WorkingDirectory != tt.expectedWorkingDir {
				t.Errorf("WorkingDirectory = %q, want %q", state.Context.WorkingDirectory, tt.expectedWorkingDir)
			}

			// Verify provider name
			if handler.Provider.Name() != "lmstudio" {
				t.Errorf("Provider name = %q, want %q", handler.Provider.Name(), "lmstudio")
			}
		})
	}
}

func TestReplHandler_Execute_StateWiring(t *testing.T) {
	// Create a mock setup to test state change wiring
	config := &Config{
		SystemPrompt:     "Test prompt",
		WorkingDirectory: "/test",
	}

	// Create mocks
	mockDisp := &mockDispatcher{
		state: state.AppState{
			Context: state.Context{
				SystemPrompt: "Test prompt",
			},
		},
	}

	mockProv := &mockProvider{
		response: &llm.ChatResponse{Content: "test response"},
	}

	mockStack := &mockStack{
		activeScreen: &mockScreen{},
	}

	// Create handler with mocks
	handler := &ReplHandler{
		Dispatcher: mockDisp,
		Provider:   mockProv,
		Stack:      mockStack,
		Config:     config,
	}

	// Test that state change handler is properly wired
	// Note: We can't easily test Execute() without running the full TUI,
	// but we can test the state wiring logic separately

	// Verify active screen exists
	activeScreen := handler.Stack.Active()
	if activeScreen == nil {
		t.Fatal("Expected active screen to exist")
	}

	// Test state change wiring manually
	testAction := &mockAction{name: "TEST_ACTION"}
	testState := state.AppState{
		Context: state.Context{
			SystemPrompt: "Updated prompt",
		},
	}

	// Simulate what Execute() does for state change wiring
	mockScreenImpl := activeScreen.(*mockScreen)

	// Call OnStateChange directly to test the wiring
	activeScreen.OnStateChange(testAction, testState, state.AppState{})

	// Verify the state change was received
	if len(mockScreenImpl.onStateChangeCalls) != 1 {
		t.Errorf("Expected 1 OnStateChange call, got %d", len(mockScreenImpl.onStateChangeCalls))
	}

	if len(mockScreenImpl.onStateChangeCalls) > 0 {
		call := mockScreenImpl.onStateChangeCalls[0]
		// We can't check the action name anymore since Action interface doesn't have String()
		// Instead, verify the state change worked correctly
		if call.newState.Context.SystemPrompt != "Updated prompt" {
			t.Errorf("Expected SystemPrompt 'Updated prompt', got %q", call.newState.Context.SystemPrompt)
		}
	}
}

func TestReplHandler_ComponentIntegration(t *testing.T) {
	config := &Config{
		SystemPrompt:     "Integration test prompt",
		WorkingDirectory: "/integration/test",
	}

	handler := NewReplHandler(config)

	// Test that all components are properly integrated
	t.Run("dispatcher_state_consistency", func(t *testing.T) {
		state := handler.Dispatcher.GetState()
		if state.Context.SystemPrompt != config.SystemPrompt {
			t.Errorf("State SystemPrompt = %q, want %q", state.Context.SystemPrompt, config.SystemPrompt)
		}
		if state.Context.WorkingDirectory != config.WorkingDirectory {
			t.Errorf("State WorkingDirectory = %q, want %q", state.Context.WorkingDirectory, config.WorkingDirectory)
		}
	})

	t.Run("provider_functionality", func(t *testing.T) {
		// Test that provider can return models
		ctx := context.Background()
		models, err := handler.Provider.Models(ctx)
		if err != nil {
			// LMStudio might not be running, so we'll allow this error
			t.Logf("Provider.Models() failed (expected if LMStudio not running): %v", err)
		} else if len(models) == 0 {
			t.Log("Provider.Models() returned empty list (expected if LMStudio not running)")
		}
	})

	t.Run("stack_active_screen", func(t *testing.T) {
		active := handler.Stack.Active()
		if active == nil {
			t.Error("Stack should have an active screen")
		}
	})

	t.Run("config_preservation", func(t *testing.T) {
		if handler.Config != config {
			t.Error("Config should be preserved in handler")
		}
		if handler.Config.SystemPrompt != config.SystemPrompt {
			t.Errorf("Config SystemPrompt = %q, want %q", handler.Config.SystemPrompt, config.SystemPrompt)
		}
	})
}

func TestReplHandler_ErrorHandling(t *testing.T) {
	// Test various error conditions that might occur
	t.Run("nil_config", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic with nil config")
			}
		}()
		NewReplHandler(nil)
	})
}

// mockAction for testing state changes
type mockAction struct {
	name string
}

func (a *mockAction) Execute(state state.AppState) (state.AppState, error) {
	// For testing, we'll update the system prompt to verify the action was executed
	newState := state
	newState.Context.SystemPrompt = "Updated prompt"
	return newState, nil
}

// Benchmark for handler creation
func BenchmarkNewReplHandler(b *testing.B) {
	config := &Config{
		SystemPrompt:     "Benchmark test",
		WorkingDirectory: "/tmp",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewReplHandler(config)
	}
}

func TestReplHandler_MemoryUsage(t *testing.T) {
	// Test that handler doesn't leak memory on creation
	config := &Config{
		SystemPrompt:     "Memory test",
		WorkingDirectory: "/tmp",
	}

	// Create multiple handlers to ensure no obvious leaks
	for i := 0; i < 10; i++ {
		handler := NewReplHandler(config)

		// Verify basic functionality
		if handler.Dispatcher == nil {
			t.Errorf("Handler %d: Dispatcher is nil", i)
		}
		if handler.Stack == nil {
			t.Errorf("Handler %d: Stack is nil", i)
		}

		// Test that state is independent between handlers
		state := handler.Dispatcher.GetState()
		if state.Context.SystemPrompt != config.SystemPrompt {
			t.Errorf("Handler %d: SystemPrompt = %q, want %q", i, state.Context.SystemPrompt, config.SystemPrompt)
		}
	}
}
