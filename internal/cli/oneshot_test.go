package cli

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/adamveld12/tai/internal/llm"
	"github.com/adamveld12/tai/internal/state"
)

// mockProvider is a mock implementation of llm.Provider for testing
type mockProvider struct {
	response *llm.ChatResponse
	err      error
	called   bool
	request  llm.ChatRequest
}

func (m *mockProvider) Name() state.SupportedProvider {
	return state.SupportedProvider("mock")
}

func (m *mockProvider) ChatCompletion(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	m.called = true
	m.request = req
	return m.response, m.err
}

func (m *mockProvider) StreamChatCompletion(ctx context.Context, req llm.ChatRequest) (<-chan llm.ChatStreamChunk, error) {
	return nil, errors.New("not implemented")
}

func (m *mockProvider) Models(ctx context.Context) ([]string, error) {
	return []string{"mock-model"}, nil
}

func (m *mockProvider) Model() string {
	return "mock-model"
}

// mockDispatcher is a mock implementation of state.Dispatcher for testing
type mockDispatcher struct {
	state state.AppState
}

func (m *mockDispatcher) GetState() state.AppState {
	return m.state
}

func (m *mockDispatcher) OnStateChange(state.OnStateChangeHandler) {
}

func (m *mockDispatcher) Dispatch(state.Action) {
}

func TestOneShotHandler_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		stdinInput     string
		mockResponse   *llm.ChatResponse
		mockError      error
		expectError    bool
		expectedOutput string
		setupStdin     func() (*os.File, func())
	}{
		{
			name: "successful execution with prompt only",
			config: &Config{
				Prompt:           "Hello AI",
				WorkingDirectory: "/tmp",
			},
			mockResponse: &llm.ChatResponse{
				Content: "Hello! How can I help you?",
			},
			expectedOutput: "Hello! How can I help you?\n",
			setupStdin: func() (*os.File, func()) {
				// Simulate terminal mode (no piped input)
				return os.Stdin, func() {}
			},
		},
		{
			name: "successful execution with stdin input",
			config: &Config{
				Prompt:           "Analyze this:",
				WorkingDirectory: "/tmp",
			},
			stdinInput: "Some input text from stdin",
			mockResponse: &llm.ChatResponse{
				Content: "Analysis complete",
			},
			expectedOutput: "Analysis complete\n",
			setupStdin: func() (*os.File, func()) {
				// Create a pipe to simulate stdin input
				r, w, _ := os.Pipe()
				go func() {
					_, _ = w.WriteString("Some input text from stdin")
					w.Close()
				}()
				return r, func() { r.Close() }
			},
		},
		{
			name: "empty prompt and no stdin input",
			config: &Config{
				Prompt:           "",
				WorkingDirectory: "/tmp",
			},
			expectError: false,
			setupStdin: func() (*os.File, func()) {
				return os.Stdin, func() {}
			},
		},
		{
			name: "LLM provider returns error",
			config: &Config{
				Prompt:           "Test prompt",
				WorkingDirectory: "/tmp",
			},
			mockError:   errors.New("LLM service unavailable"),
			expectError: true,
			setupStdin: func() (*os.File, func()) {
				return os.Stdin, func() {}
			},
		},
		{
			name: "stdin input only without prompt",
			config: &Config{
				Prompt:           "",
				WorkingDirectory: "/tmp",
			},
			stdinInput: "Direct stdin input",
			mockResponse: &llm.ChatResponse{
				Content: "Response to stdin",
			},
			expectedOutput: "Response to stdin\n",
			setupStdin: func() (*os.File, func()) {
				r, w, _ := os.Pipe()
				go func() {
					_, _ = w.WriteString("Direct stdin input")
					w.Close()
				}()
				return r, func() { r.Close() }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup stdin
			oldStdin := os.Stdin
			stdin, cleanup := tt.setupStdin()
			os.Stdin = stdin
			defer func() {
				cleanup()
				os.Stdin = oldStdin
			}()

			// Create mock provider
			mockProv := &mockProvider{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			// Create mock dispatcher
			mockDisp := &mockDispatcher{
				state: state.AppState{
					Context: state.Context{
						SystemPrompt: "test system prompt",
					},
				},
			}

			// Create handler with mocks
			handler := &OneShotHandler{
				Dispatcher: mockDisp,
				Provider:   mockProv,
				config:     tt.config,
			}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute
			err := handler.Execute()

			// Restore stdout
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = oldStdout

			// Check error
			if (err != nil) != tt.expectError {
				t.Errorf("Execute() error = %v, expectError %v", err, tt.expectError)
			}

			// Check output
			if !tt.expectError && string(out) != tt.expectedOutput {
				t.Errorf("Execute() output = %q, want %q", string(out), tt.expectedOutput)
			}

			// Verify provider was called when expected
			if !tt.expectError && tt.config.Prompt != "" || tt.stdinInput != "" {
				if !mockProv.called {
					t.Error("Expected provider.ChatCompletion to be called")
				}
			}
		})
	}
}

func TestOneShotHandler_readFromStdin(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		setupStdin  func() (*os.File, func())
		expected    string
		expectError bool
	}{
		{
			name: "read multiple lines from stdin",
			input: `line1
line2
line3`,
			setupStdin: func() (*os.File, func()) {
				r, w, _ := os.Pipe()
				go func() {
					_, _ = w.WriteString("line1\nline2\nline3")
					w.Close()
				}()
				return r, func() { r.Close() }
			},
			expected: "line1\nline2\nline3",
		},
		{
			name: "terminal mode returns empty string",
			setupStdin: func() (*os.File, func()) {
				// Use actual stdin which is in terminal mode during tests
				return os.Stdin, func() {}
			},
			expected: "",
		},
		{
			name:  "single line without newline",
			input: "single line",
			setupStdin: func() (*os.File, func()) {
				r, w, _ := os.Pipe()
				go func() {
					_, _ = w.WriteString("single line")
					w.Close()
				}()
				return r, func() { r.Close() }
			},
			expected: "single line",
		},
		{
			name:  "empty stdin",
			input: "",
			setupStdin: func() (*os.File, func()) {
				r, w, _ := os.Pipe()
				go func() {
					w.Close()
				}()
				return r, func() { r.Close() }
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup stdin
			oldStdin := os.Stdin
			stdin, cleanup := tt.setupStdin()
			os.Stdin = stdin
			defer func() {
				cleanup()
				os.Stdin = oldStdin
			}()

			handler := &OneShotHandler{}
			result, err := handler.readFromStdin()

			if (err != nil) != tt.expectError {
				t.Errorf("readFromStdin() error = %v, expectError %v", err, tt.expectError)
			}

			if result != tt.expected {
				t.Errorf("readFromStdin() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestOneShotHandler_MessageConstruction(t *testing.T) {
	tests := []struct {
		name        string
		prompt      string
		stdinInput  string
		expectedMsg string
	}{
		{
			name:        "prompt and stdin combined",
			prompt:      "Analyze this:",
			stdinInput:  "data to analyze",
			expectedMsg: "Analyze this:\ndata to analyze",
		},
		{
			name:        "prompt only",
			prompt:      "Hello AI",
			stdinInput:  "",
			expectedMsg: "Hello AI",
		},
		{
			name:        "stdin only",
			prompt:      "",
			stdinInput:  "Just stdin content",
			expectedMsg: "Just stdin content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup stdin with input
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			if tt.stdinInput != "" {
				go func() {
					_, _ = w.WriteString(tt.stdinInput)
					w.Close()
				}()
			} else {
				w.Close()
			}
			os.Stdin = r
			defer func() {
				r.Close()
				os.Stdin = oldStdin
			}()

			// Create mock provider to capture the request
			mockProv := &mockProvider{
				response: &llm.ChatResponse{Content: "response"},
			}

			// Create mock dispatcher
			mockDisp := &mockDispatcher{
				state: state.AppState{
					Context: state.Context{},
				},
			}

			handler := &OneShotHandler{
				Dispatcher: mockDisp,
				Provider:   mockProv,
				config: &Config{
					Prompt:           tt.prompt,
					WorkingDirectory: "/tmp",
				},
			}

			// Redirect stdout to avoid test output noise
			oldStdout := os.Stdout
			_, wOut, _ := os.Pipe()
			os.Stdout = wOut
			defer func() {
				wOut.Close()
				os.Stdout = oldStdout
			}()

			_ = handler.Execute()

			// Check the message content
			if mockProv.called && len(mockProv.request.Messages) > 0 {
				actualMsg := mockProv.request.Messages[0].Content
				if actualMsg != tt.expectedMsg {
					t.Errorf("Message content = %q, want %q", actualMsg, tt.expectedMsg)
				}
			}
		})
	}
}
