package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adamveld12/tai/internal/state"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Constants and Configuration
// =============================================================================

const (
	// testTimeout is short enough to not slow tests but long enough
	// to work on slow CI systems
	testTimeout = 5 * time.Second

	// shortTimeout for context cancellation tests
	shortTimeout = 100 * time.Millisecond

	// streamChunkDelay simulates realistic streaming delays
	streamChunkDelay = 10 * time.Millisecond

	// goroutineCleanupWait allows time for goroutines to clean up
	goroutineCleanupWait = 200 * time.Millisecond
)

// =============================================================================
// Test Infrastructure
// =============================================================================

// newTestProvider creates a provider for testing with proper error handling
func newTestProvider(t *testing.T, config ProviderConfig) *OpenAIProvider {
	t.Helper()

	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:1234/v1"
	}
	if config.APIKey == "" {
		config.APIKey = "test-key"
	}

	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err, "failed to create provider")
	require.NotNil(t, provider, "provider should not be nil")

	return provider
}

// recordedRequest captures details about HTTP requests for verification
type recordedRequest struct {
	Method    string
	Path      string
	Body      []byte
	Headers   http.Header
	Timestamp time.Time
}

// mockResponse defines how the mock server should respond
type mockResponse struct {
	StatusCode int
	Body       interface{}
	Delay      time.Duration
	Error      error
	Headers    map[string]string
}

// mockServer provides controlled HTTP responses with request tracking
type mockServer struct {
	server    *httptest.Server
	requests  []recordedRequest
	responses []mockResponse
	mu        sync.RWMutex
	callCount int
	t         *testing.T
}

// newMockServer creates a new mock server that records requests and serves configured responses
func newMockServer(t *testing.T, responses ...mockResponse) *mockServer {
	t.Helper()

	ms := &mockServer{
		responses: responses,
		t:         t,
	}

	ms.server = httptest.NewServer(http.HandlerFunc(ms.handleRequest))

	return ms
}

func (ms *mockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Record the request
	body := make([]byte, 0)
	if r.Body != nil {
		bodyBytes := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(bodyBytes)
		body = bodyBytes
	}

	ms.requests = append(ms.requests, recordedRequest{
		Method:    r.Method,
		Path:      r.URL.Path,
		Body:      body,
		Headers:   r.Header.Clone(),
		Timestamp: time.Now(),
	})

	// Check if we have a response configured
	if ms.callCount >= len(ms.responses) {
		ms.t.Errorf("unexpected request #%d: %s %s", ms.callCount+1, r.Method, r.URL.Path)
		http.Error(w, "no more mock responses configured", http.StatusInternalServerError)
		return
	}

	resp := ms.responses[ms.callCount]
	ms.callCount++

	// Add delay if specified
	if resp.Delay > 0 {
		time.Sleep(resp.Delay)
	}

	// Set custom headers
	for key, value := range resp.Headers {
		w.Header().Set(key, value)
	}

	// Handle error responses
	if resp.Error != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		if resp.StatusCode == 0 {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"message": resp.Error.Error(),
				"type":    "api_error",
			},
		})
		return
	}

	// Handle success responses
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	if resp.StatusCode == 0 {
		w.WriteHeader(http.StatusOK)
	}

	if resp.Body != nil {
		err := json.NewEncoder(w).Encode(resp.Body)
		if err != nil {
			ms.t.Errorf("failed to encode response: %v", err)
		}
	}
}

func (ms *mockServer) URL() string {
	return ms.server.URL
}

func (ms *mockServer) Close() {
	ms.server.Close()
}

// GetRequests returns all recorded requests (thread-safe)
func (ms *mockServer) GetRequests() []recordedRequest {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	requests := make([]recordedRequest, len(ms.requests))
	copy(requests, ms.requests)
	return requests
}

// RequestCount returns the number of requests made
func (ms *mockServer) RequestCount() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.requests)
}

// =============================================================================
// Constructor Tests
// =============================================================================

// TestNewLMStudioProvider verifies provider creation with various configurations.
// This ensures proper default handling and configuration validation.
func TestNewOpenAIProvider(t *testing.T) {
	tests := []struct {
		name   string
		config ProviderConfig
		verify func(t *testing.T, p *OpenAIProvider)
	}{
		{
			name:   "defaults_applied_correctly",
			config: ProviderConfig{},
			verify: func(t *testing.T, p *OpenAIProvider) {
				assert.Equal(t, "http://localhost:1234/v1", p.config.BaseURL)
				assert.Equal(t, "test-key", p.config.APIKey) // Override from newTestProvider
				assert.Equal(t, "gemma-3n-e4b-it", p.config.DefaultModel)
				assert.Equal(t, 300*time.Second, p.config.Timeout)
				assert.Equal(t, 0, p.config.MaxRetries) // Default is 0, retryRequest uses 3 if 0
			},
		},
		{
			name: "custom_values_preserved",
			config: ProviderConfig{
				BaseURL:      "http://custom.example.com/v1",
				APIKey:       "custom-key",
				DefaultModel: "custom-model",
				Timeout:      30 * time.Second,
				MaxRetries:   5,
			},
			verify: func(t *testing.T, p *OpenAIProvider) {
				assert.Equal(t, "http://custom.example.com/v1", p.config.BaseURL)
				assert.Equal(t, "custom-key", p.config.APIKey)
				assert.Equal(t, "custom-model", p.config.DefaultModel)
				assert.Equal(t, 30*time.Second, p.config.Timeout)
				assert.Equal(t, 5, p.config.MaxRetries)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := newTestProvider(t, tt.config)
			require.NotNil(t, provider.client, "client should be initialized")
			tt.verify(t, provider)
		})
	}
}

// =============================================================================
// ChatCompletion Tests
// =============================================================================

// TestChatCompletion_SuccessScenarios verifies the ChatCompletion method handles
// various successful request types correctly. This covers the main user workflows.
func TestOpenAIProvider_ChatCompletion_SuccessScenarios(t *testing.T) {
	tests := []struct {
		name        string
		description string
		request     ChatRequest
		response    openai.ChatCompletionResponse
		verify      func(t *testing.T, resp *ChatResponse, err error)
	}{
		{
			name:        "basic_conversation",
			description: "Simple user message should work correctly",
			request: ChatRequest{
				Messages: []state.Message{
					{Role: state.RoleUser, Content: "Hello, how are you?"},
				},
			},
			response: openai.ChatCompletionResponse{
				Model:   "test-model",
				Created: 1234567890,
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "I'm doing well, thank you!",
						},
						FinishReason: openai.FinishReasonStop,
					},
				},
				Usage: openai.Usage{
					PromptTokens:     5,
					CompletionTokens: 6,
					TotalTokens:      11,
				},
			},
			verify: func(t *testing.T, resp *ChatResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "test-model", resp.Model)
				assert.Equal(t, "I'm doing well, thank you!", resp.Content)
				assert.Equal(t, "stop", resp.FinishReason)
				assert.Equal(t, 5, resp.Usage.PromptTokens)
				assert.Equal(t, 6, resp.Usage.CompletionTokens)
				assert.Equal(t, 11, resp.Usage.TotalTokens)
				assert.True(t, resp.Duration > 0, "duration should be measured")
			},
		},
		{
			name:        "conversation_with_system_prompt",
			description: "System prompts should be properly prepended to messages",
			request: ChatRequest{
				SystemPrompt: "You are a helpful assistant",
				Messages: []state.Message{
					{Role: state.RoleUser, Content: "What is 2+2?"},
				},
			},
			response: openai.ChatCompletionResponse{
				Model:   "test-model",
				Created: 1234567890,
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "2+2 equals 4",
						},
						FinishReason: openai.FinishReasonStop,
					},
				},
			},
			verify: func(t *testing.T, resp *ChatResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "2+2 equals 4", resp.Content)
			},
		},
		{
			name:        "function_calling_workflow",
			description: "Tool calls should be properly converted and returned",
			request: ChatRequest{
				Messages: []state.Message{
					{Role: state.RoleUser, Content: "What's the weather in New York?"},
				},
				Tools: []Tool{
					{
						Type: "function",
						Function: ToolFunction{
							Name:        "get_weather",
							Description: "Get current weather",
							Parameters: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"location": map[string]interface{}{
										"type": "string",
									},
								},
							},
						},
					},
				},
				ToolChoice: "auto",
			},
			response: openai.ChatCompletionResponse{
				Model:   "test-model",
				Created: 1234567890,
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "I'll check the weather for you.",
							ToolCalls: []openai.ToolCall{
								{
									ID:   "call_123",
									Type: openai.ToolTypeFunction,
									Function: openai.FunctionCall{
										Name:      "get_weather",
										Arguments: `{"location": "New York"}`,
									},
								},
							},
						},
						FinishReason: openai.FinishReasonToolCalls,
					},
				},
			},
			verify: func(t *testing.T, resp *ChatResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "I'll check the weather for you.", resp.Content)
				assert.Equal(t, "tool_calls", resp.FinishReason)
				require.Len(t, resp.ToolCalls, 1)

				toolCall := resp.ToolCalls[0]
				assert.Equal(t, "call_123", toolCall.ID)
				assert.Equal(t, "function", toolCall.Type)
				assert.Equal(t, "get_weather", toolCall.Function.Name)
				assert.Equal(t, `{"location": "New York"}`, toolCall.Function.Arguments)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockServer(t, mockResponse{
				StatusCode: http.StatusOK,
				Body:       tt.response,
			})
			defer mock.Close()

			provider := newTestProvider(t, ProviderConfig{
				BaseURL: mock.URL(),
				Timeout: testTimeout,
			})

			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			resp, err := provider.ChatCompletion(ctx, tt.request)
			tt.verify(t, resp, err)

			// Verify exactly one request was made
			assert.Equal(t, 1, mock.RequestCount(), "should make exactly one HTTP request")
		})
	}
}

func TestOpenAIProvider_ChatCompletion_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name             string
		description      string
		request          ChatRequest
		responses        []mockResponse
		expectedReqCount int
		verify           func(t *testing.T, resp *ChatResponse, err error)
	}{
		{
			name:        "non_retryable_error_invalid_api_key",
			description: "Invalid API key errors should not be retried",
			request: ChatRequest{
				Messages: []state.Message{
					{Role: state.RoleUser, Content: "Hello"},
				},
			},
			responses: []mockResponse{
				{
					StatusCode: http.StatusUnauthorized,
					Error:      errors.New("invalid_api_key"),
				},
			},
			expectedReqCount: 1, // Should not retry invalid_api_key errors
			verify: func(t *testing.T, resp *ChatResponse, err error) {
				require.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "chat completion failed")
			},
		},
		{
			name:        "non_retryable_error_model_not_found",
			description: "Model not found errors should not be retried",
			request: ChatRequest{
				Messages: []state.Message{
					{Role: state.RoleUser, Content: "Hello"},
				},
			},
			responses: []mockResponse{
				{
					StatusCode: http.StatusNotFound,
					Error:      errors.New("model_not_found"),
				},
			},
			expectedReqCount: 1, // Should not retry model_not_found errors
			verify: func(t *testing.T, resp *ChatResponse, err error) {
				require.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "chat completion failed")
			},
		},
		{
			name:        "retryable_error_general",
			description: "General errors (400s, 500s) should be retried",
			request: ChatRequest{
				Messages: []state.Message{
					{Role: state.RoleUser, Content: "Hello"},
				},
			},
			responses: []mockResponse{
				{StatusCode: http.StatusBadRequest, Error: errors.New("Bad request")},
				{StatusCode: http.StatusBadRequest, Error: errors.New("Bad request")},
				{StatusCode: http.StatusBadRequest, Error: errors.New("Bad request")},
			},
			expectedReqCount: 3, // Should retry general errors
			verify: func(t *testing.T, resp *ChatResponse, err error) {
				require.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "request failed after 3 retries")
			},
		},
		{
			name:        "http_500_with_retries",
			description: "Server errors should be retried with exponential backoff",
			request: ChatRequest{
				Messages: []state.Message{
					{Role: state.RoleUser, Content: "Hello"},
				},
			},
			responses: []mockResponse{
				{StatusCode: http.StatusInternalServerError, Error: errors.New("Server error")},
				{StatusCode: http.StatusInternalServerError, Error: errors.New("Server error")},
				{StatusCode: http.StatusInternalServerError, Error: errors.New("Server error")},
			},
			expectedReqCount: 3, // Should retry 3 times (default MaxRetries)
			verify: func(t *testing.T, resp *ChatResponse, err error) {
				require.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "request failed after 3 retries")
			},
		},
		{
			name:        "http_429_with_retries",
			description: "Rate limiting should be retried",
			request: ChatRequest{
				Messages: []state.Message{
					{Role: state.RoleUser, Content: "Hello"},
				},
			},
			responses: []mockResponse{
				{StatusCode: http.StatusTooManyRequests, Error: errors.New("Rate limit exceeded")},
				{StatusCode: http.StatusTooManyRequests, Error: errors.New("Rate limit exceeded")},
				{StatusCode: http.StatusTooManyRequests, Error: errors.New("Rate limit exceeded")},
			},
			expectedReqCount: 3,
			verify: func(t *testing.T, resp *ChatResponse, err error) {
				require.Error(t, err)
				assert.Nil(t, resp)
			},
		},
		{
			name:        "empty_response_choices",
			description: "Empty choices array should not crash the application",
			request: ChatRequest{
				Messages: []state.Message{
					{Role: state.RoleUser, Content: "Hello"},
				},
			},
			responses: []mockResponse{
				{
					StatusCode: http.StatusOK,
					Body: openai.ChatCompletionResponse{
						Model:   "test-model",
						Created: 1234567890,
						Choices: []openai.ChatCompletionChoice{}, // Empty choices
					},
				},
			},
			expectedReqCount: 1,
			verify: func(t *testing.T, resp *ChatResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "", resp.Content)
				assert.Equal(t, "", resp.FinishReason)
				assert.Len(t, resp.ToolCalls, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockServer(t, tt.responses...)
			defer mock.Close()

			provider := newTestProvider(t, ProviderConfig{
				BaseURL: mock.URL(),
				Timeout: testTimeout,
			})

			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			resp, err := provider.ChatCompletion(ctx, tt.request)
			tt.verify(t, resp, err)

			// Verify the expected number of requests were made
			assert.Equal(t, tt.expectedReqCount, mock.RequestCount(),
				"unexpected number of HTTP requests")
		})
	}
}

// TestChatCompletion_ContextCancellation verifies proper context handling.
// This is critical for preventing hung requests in production.
func TestOpenAIProvider_ChatCompletion_ContextCancellation(t *testing.T) {
	// Create a mock that delays longer than our context timeout
	mock := newMockServer(t, mockResponse{
		StatusCode: http.StatusOK,
		Delay:      testTimeout, // Longer than our context
		Body: openai.ChatCompletionResponse{
			Model: "test-model",
		},
	})
	defer mock.Close()

	provider := newTestProvider(t, ProviderConfig{
		BaseURL: mock.URL(),
		Timeout: testTimeout,
	})

	// Use a short timeout to trigger cancellation
	ctx, cancel := context.WithTimeout(context.Background(), shortTimeout)
	defer cancel()

	resp, err := provider.ChatCompletion(ctx, ChatRequest{
		Messages: []state.Message{
			{Role: state.RoleUser, Content: "Hello"},
		},
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "context")
}

// =============================================================================
// StreamChatCompletion Tests
// =============================================================================

// streamingMockServer creates a specialized server for streaming tests
type streamingMockServer struct {
	server *httptest.Server
	chunks []string
	delay  time.Duration
	t      *testing.T
}

func newStreamingMockServer(t *testing.T, chunks []string, delay time.Duration) *streamingMockServer {
	t.Helper()

	sms := &streamingMockServer{
		chunks: chunks,
		delay:  delay,
		t:      t,
	}

	sms.server = httptest.NewServer(http.HandlerFunc(sms.handleStreamRequest))
	return sms
}

func (sms *streamingMockServer) handleStreamRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		sms.t.Error("ResponseWriter does not support flushing")
		return
	}

	for _, chunk := range sms.chunks {
		if sms.delay > 0 {
			time.Sleep(sms.delay)
		}

		_, err := w.Write([]byte(chunk + "\n"))
		if err != nil {
			sms.t.Logf("Error writing chunk: %v", err)
			return
		}
		flusher.Flush()
	}
}

func (sms *streamingMockServer) URL() string {
	return sms.server.URL
}

func (sms *streamingMockServer) Close() {
	sms.server.Close()
}

// TestStreamChatCompletion_SuccessScenarios verifies streaming functionality works correctly.
// Streaming is critical for providing real-time user feedback in chat applications.
func TestOpenAIProvider_StreamChatCompletion_SuccessScenarios(t *testing.T) {
	tests := []struct {
		name        string
		description string
		request     ChatRequest
		chunks      []string
		verify      func(t *testing.T, chunks []ChatStreamChunk, err error)
	}{
		{
			name:        "basic_streaming_conversation",
			description: "Simple streaming should work with multiple content chunks",
			request: ChatRequest{
				Messages: []state.Message{
					{Role: state.RoleUser, Content: "Tell me a short story"},
				},
			},
			chunks: []string{
				`data: {"choices":[{"delta":{"content":"Once"},"finish_reason":null}],"usage":{"prompt_tokens":10,"completion_tokens":1,"total_tokens":11}}`,
				`data: {"choices":[{"delta":{"content":" upon"},"finish_reason":null}],"usage":{"prompt_tokens":10,"completion_tokens":2,"total_tokens":12}}`,
				`data: {"choices":[{"delta":{"content":" a"},"finish_reason":null}],"usage":{"prompt_tokens":10,"completion_tokens":3,"total_tokens":13}}`,
				`data: {"choices":[{"delta":{"content":" time."},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":4,"total_tokens":14}}`,
				`data: [DONE]`,
			},
			verify: func(t *testing.T, chunks []ChatStreamChunk, err error) {
				require.NoError(t, err)

				// Filter content chunks (ignore done-only chunks)
				contentChunks := []ChatStreamChunk{}
				var doneChunk *ChatStreamChunk

				for _, chunk := range chunks {
					if chunk.Done && chunk.Delta == "" {
						doneChunk = &chunk
					} else if chunk.Delta != "" {
						contentChunks = append(contentChunks, chunk)
					}
				}

				require.Len(t, contentChunks, 4, "should receive 4 content chunks")
				assert.Equal(t, "Once", contentChunks[0].Delta)
				assert.Equal(t, " upon", contentChunks[1].Delta)
				assert.Equal(t, " a", contentChunks[2].Delta)
				assert.Equal(t, " time.", contentChunks[3].Delta)

				require.NotNil(t, doneChunk, "should receive done signal")
				assert.True(t, doneChunk.Done)
			},
		},
		{
			name:        "streaming_with_tool_calls",
			description: "Tool calls in streaming should be properly parsed and assembled",
			request: ChatRequest{
				Messages: []state.Message{
					{Role: state.RoleUser, Content: "What's the weather?"},
				},
				Tools: []Tool{
					{
						Type: "function",
						Function: ToolFunction{
							Name: "get_weather",
						},
					},
				},
			},
			chunks: []string{
				`data: {"choices":[{"delta":{"content":"I'll check the weather","tool_calls":[{"id":"call_123","type":"function","function":{"name":"get_weather","arguments":"{\"location\":"}}]},"finish_reason":null}],"usage":{"prompt_tokens":15,"completion_tokens":5,"total_tokens":20}}`,
				`data: {"choices":[{"delta":{"tool_calls":[{"function":{"arguments":"\"New York\"}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":15,"completion_tokens":10,"total_tokens":25}}`,
				`data: [DONE]`,
			},
			verify: func(t *testing.T, chunks []ChatStreamChunk, err error) {
				require.NoError(t, err)
				require.GreaterOrEqual(t, len(chunks), 1)

				// Check that we received tool calls
				hasToolCall := false
				for _, chunk := range chunks {
					if len(chunk.ToolCalls) > 0 {
						hasToolCall = true
						assert.Equal(t, "call_123", chunk.ToolCalls[0].ID)
						assert.Equal(t, "function", chunk.ToolCalls[0].Type)
						assert.Equal(t, "get_weather", chunk.ToolCalls[0].Function.Name)
						// Note: In real streaming, arguments would be assembled across chunks
						break
					}
				}
				assert.True(t, hasToolCall, "should receive tool calls in streaming")
			},
		},
		{
			name:        "empty_stream",
			description: "Empty streams should complete gracefully without error",
			request: ChatRequest{
				Messages: []state.Message{
					{Role: state.RoleUser, Content: "Hello"},
				},
			},
			chunks: []string{
				`data: [DONE]`,
			},
			verify: func(t *testing.T, chunks []ChatStreamChunk, err error) {
				require.NoError(t, err)

				// Should only receive a done chunk
				require.Len(t, chunks, 1)
				assert.True(t, chunks[0].Done)
				assert.Equal(t, "", chunks[0].Delta)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newStreamingMockServer(t, tt.chunks, streamChunkDelay)
			defer mock.Close()

			provider := newTestProvider(t, ProviderConfig{
				BaseURL: mock.URL(),
				Timeout: testTimeout,
			})

			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			chunkChan, err := provider.StreamChatCompletion(ctx, tt.request)
			require.NoError(t, err, "StreamChatCompletion should not error immediately")

			var receivedChunks []ChatStreamChunk
			for chunk := range chunkChan {
				receivedChunks = append(receivedChunks, chunk)
			}

			tt.verify(t, receivedChunks, nil)
		})
	}
}

// TestStreamChatCompletion_ErrorScenarios verifies proper error handling in streaming.
// Error handling in streaming is complex because errors can occur at different stages.
func TestOpenAIProvider_StreamChatCompletion_ErrorScenarios(t *testing.T) {
	t.Run("connection_error_before_stream", func(t *testing.T) {
		// Use an invalid URL to trigger immediate connection error
		provider := newTestProvider(t, ProviderConfig{
			BaseURL: "http://invalid-host-that-does-not-exist.example.com",
			Timeout: shortTimeout,
		})

		ctx, cancel := context.WithTimeout(context.Background(), shortTimeout)
		defer cancel()

		chunkChan, err := provider.StreamChatCompletion(ctx, ChatRequest{
			Messages: []state.Message{
				{Role: state.RoleUser, Content: "Hello"},
			},
		})

		// Should get immediate error for connection failure
		require.Error(t, err)
		assert.Nil(t, chunkChan)
		assert.Contains(t, err.Error(), "stream creation failed")
	})

	t.Run("context_cancellation_during_stream", func(t *testing.T) {
		// Create a mock that sends many chunks with delays to ensure cancellation happens mid-stream
		chunks := []string{
			`data: {"choices":[{"delta":{"content":"Starting"},"finish_reason":null}],"usage":{"prompt_tokens":5,"completion_tokens":1,"total_tokens":6}}`,
			`data: {"choices":[{"delta":{"content":" a"},"finish_reason":null}],"usage":{"prompt_tokens":5,"completion_tokens":2,"total_tokens":7}}`,
			`data: {"choices":[{"delta":{"content":" very"},"finish_reason":null}],"usage":{"prompt_tokens":5,"completion_tokens":3,"total_tokens":8}}`,
			`data: {"choices":[{"delta":{"content":" long"},"finish_reason":null}],"usage":{"prompt_tokens":5,"completion_tokens":4,"total_tokens":9}}`,
			`data: {"choices":[{"delta":{"content":" message"},"finish_reason":null}],"usage":{"prompt_tokens":5,"completion_tokens":5,"total_tokens":10}}`,
			`data: {"choices":[{"delta":{"content":" that"},"finish_reason":null}],"usage":{"prompt_tokens":5,"completion_tokens":6,"total_tokens":11}}`,
			`data: {"choices":[{"delta":{"content":" should"},"finish_reason":null}],"usage":{"prompt_tokens":5,"completion_tokens":7,"total_tokens":12}}`,
			`data: {"choices":[{"delta":{"content":" be"},"finish_reason":null}],"usage":{"prompt_tokens":5,"completion_tokens":8,"total_tokens":13}}`,
			`data: {"choices":[{"delta":{"content":" cancelled"},"finish_reason":null}],"usage":{"prompt_tokens":5,"completion_tokens":9,"total_tokens":14}}`,
			`data: {"choices":[{"delta":{"content":" midway"},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":10,"total_tokens":15}}`,
			`data: [DONE]`,
		}

		mock := newStreamingMockServer(t, chunks, 100*time.Millisecond) // Long enough delay
		defer mock.Close()

		provider := newTestProvider(t, ProviderConfig{
			BaseURL: mock.URL(),
			Timeout: testTimeout, // Long provider timeout
		})

		// Use a context that will be cancelled during streaming
		ctx, cancel := context.WithCancel(context.Background())

		chunkChan, err := provider.StreamChatCompletion(ctx, ChatRequest{
			Messages: []state.Message{
				{Role: state.RoleUser, Content: "Hello"},
			},
		})

		require.NoError(t, err, "initial stream creation should succeed")
		require.NotNil(t, chunkChan)

		// Cancel context after allowing only a couple chunks through
		go func() {
			time.Sleep(250 * time.Millisecond) // Allow 2-3 chunks
			cancel()
		}()

		// Collect chunks until channel closes
		var receivedChunks []ChatStreamChunk
		for chunk := range chunkChan {
			receivedChunks = append(receivedChunks, chunk)
		}

		// Should have received some but not all chunks
		assert.Greater(t, len(receivedChunks), 0, "should receive some chunks before cancellation")
		assert.Less(t, len(receivedChunks), len(chunks), "should not receive all chunks due to cancellation")

		// Verify that we didn't receive the full completion sequence
		totalContent := ""
		for _, chunk := range receivedChunks {
			totalContent += chunk.Delta
		}

		// Should not contain the full message since it was cancelled
		assert.NotContains(t, totalContent, "cancelled midway",
			"stream was cancelled before completion")
	})
}

// TestStreamChatCompletion_GoroutineCleanup verifies that streaming doesn't leak goroutines.
// This is critical for production systems that handle many concurrent streams.
func TestOpenAIProvider_StreamChatCompletion_GoroutineCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping goroutine cleanup test in short mode")
	}

	initialGoroutines := runtime.NumGoroutine()

	// Create multiple short streams to test cleanup
	for i := 0; i < 5; i++ {
		func() {
			chunks := []string{
				`data: {"choices":[{"delta":{"content":"Hello"},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":1,"total_tokens":6}}`,
				`data: [DONE]`,
			}

			mock := newStreamingMockServer(t, chunks, streamChunkDelay)
			defer mock.Close()

			provider := newTestProvider(t, ProviderConfig{
				BaseURL: mock.URL(),
				Timeout: testTimeout,
			})

			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			chunkChan, err := provider.StreamChatCompletion(ctx, ChatRequest{
				Messages: []state.Message{
					{Role: state.RoleUser, Content: "Hello"},
				},
			})

			require.NoError(t, err)
			require.NotNil(t, chunkChan)

			// Consume all chunks
			for range chunkChan {
				// Just drain the channel
			}
		}()
	}

	// Allow time for goroutines to clean up
	time.Sleep(goroutineCleanupWait)

	finalGoroutines := runtime.NumGoroutine()

	// Allow some tolerance for test goroutines, but should not have significantly increased
	maxAcceptableIncrease := 2
	assert.LessOrEqual(t, finalGoroutines-initialGoroutines, maxAcceptableIncrease,
		"too many goroutines leaked, expected <= %d increase, got %d increase",
		maxAcceptableIncrease, finalGoroutines-initialGoroutines)
}

// =============================================================================
// Retry Logic Tests
// =============================================================================

// TestRetryLogic_ExponentialBackoff verifies that the retry mechanism:
// 1. Retries transient errors but not specific errors (invalid_api_key, model_not_found)
// 2. Uses exponential backoff to avoid overwhelming the server
// 3. Respects context cancellation immediately
// 4. Preserves the original error when all retries fail
//
// This is critical because LMStudio can have transient failures when
// loading models or under high load.
func TestOpenAIProvider_RetryLogic_ExponentialBackoff(t *testing.T) {
	tests := []struct {
		name          string
		description   string
		responses     []mockResponse
		maxRetries    int
		expectedCalls int
		expectSuccess bool
		verifyError   func(t *testing.T, err error)
		verifyBackoff bool
	}{
		{
			name:        "success_on_first_attempt",
			description: "No retries needed when first request succeeds",
			responses: []mockResponse{
				{
					StatusCode: http.StatusOK,
					Body: openai.ChatCompletionResponse{
						Choices: []openai.ChatCompletionChoice{
							{Message: openai.ChatCompletionMessage{Content: "Success"}},
						},
					},
				},
			},
			maxRetries:    3,
			expectedCalls: 1,
			expectSuccess: true,
		},
		{
			name:        "success_after_retries",
			description: "Should retry transient errors and eventually succeed",
			responses: []mockResponse{
				{StatusCode: http.StatusInternalServerError, Error: errors.New("temporary error")},
				{StatusCode: http.StatusBadGateway, Error: errors.New("bad gateway")},
				{
					StatusCode: http.StatusOK,
					Body: openai.ChatCompletionResponse{
						Choices: []openai.ChatCompletionChoice{
							{Message: openai.ChatCompletionMessage{Content: "Success after retries"}},
						},
					},
				},
			},
			maxRetries:    3,
			expectedCalls: 3,
			expectSuccess: true,
			verifyBackoff: true, // Verify exponential backoff timing
		},
		{
			name:        "no_retry_invalid_api_key",
			description: "Should not retry invalid_api_key errors",
			responses: []mockResponse{
				{StatusCode: http.StatusUnauthorized, Error: errors.New("invalid_api_key")},
			},
			maxRetries:    3,
			expectedCalls: 1,
			expectSuccess: false,
			verifyError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "invalid_api_key")
				assert.NotContains(t, err.Error(), "retries", "should not mention retries")
			},
		},
		{
			name:        "no_retry_model_not_found",
			description: "Should not retry model_not_found errors",
			responses: []mockResponse{
				{StatusCode: http.StatusNotFound, Error: errors.New("model_not_found")},
			},
			maxRetries:    3,
			expectedCalls: 1,
			expectSuccess: false,
			verifyError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "model_not_found")
				assert.NotContains(t, err.Error(), "retries", "should not mention retries")
			},
		},
		{
			name:        "exhaust_all_retries",
			description: "Should fail after exhausting all retries",
			responses: []mockResponse{
				{StatusCode: http.StatusInternalServerError, Error: errors.New("persistent error")},
				{StatusCode: http.StatusInternalServerError, Error: errors.New("persistent error")},
				{StatusCode: http.StatusInternalServerError, Error: errors.New("persistent error")},
			},
			maxRetries:    3,
			expectedCalls: 3,
			expectSuccess: false,
			verifyError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "request failed after 3 retries")
				assert.Contains(t, err.Error(), "persistent error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Now()

			mock := newMockServer(t, tt.responses...)
			defer mock.Close()

			provider := newTestProvider(t, ProviderConfig{
				BaseURL:    mock.URL(),
				MaxRetries: tt.maxRetries,
				Timeout:    testTimeout,
			})

			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			resp, err := provider.ChatCompletion(ctx, ChatRequest{
				Messages: []state.Message{{Role: state.RoleUser, Content: "Test"}},
			})

			elapsed := time.Since(startTime)

			// Verify success/failure expectation
			if tt.expectSuccess {
				require.NoError(t, err)
				require.NotNil(t, resp)
			} else {
				require.Error(t, err)
				assert.Nil(t, resp)
				if tt.verifyError != nil {
					tt.verifyError(t, err)
				}
			}

			// Verify number of calls made
			assert.Equal(t, tt.expectedCalls, mock.RequestCount(),
				"unexpected number of HTTP requests")

			// Verify exponential backoff timing if requested
			if tt.verifyBackoff && tt.expectedCalls > 1 {
				// With exponential backoff: 1s + 2s = 3s minimum for 3 calls
				expectedMinDuration := time.Duration(tt.expectedCalls-1) * time.Second
				assert.GreaterOrEqual(t, elapsed, expectedMinDuration,
					"requests completed too quickly, backoff may not be working")
			}
		})
	}
}

// TestRetryLogic_ContextCancellation verifies context cancellation during retries.
// Context cancellation should be respected immediately, even during backoff periods.
func TestOpenAIProvider_RetryLogic_ContextCancellation(t *testing.T) {
	// Create responses that would trigger retries
	responses := []mockResponse{
		{StatusCode: http.StatusInternalServerError, Error: errors.New("first error")},
		{StatusCode: http.StatusInternalServerError, Error: errors.New("second error")},
		{StatusCode: http.StatusInternalServerError, Error: errors.New("third error")},
	}

	mock := newMockServer(t, responses...)
	defer mock.Close()

	provider := newTestProvider(t, ProviderConfig{
		BaseURL:    mock.URL(),
		MaxRetries: 5, // Lots of retries to ensure cancellation happens first
		Timeout:    testTimeout,
	})

	// Create a context that will be cancelled during the backoff period
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	startTime := time.Now()

	resp, err := provider.ChatCompletion(ctx, ChatRequest{
		Messages: []state.Message{{Role: state.RoleUser, Content: "Test"}},
	})

	elapsed := time.Since(startTime)

	// Should fail with context error
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "context")

	// Should have cancelled before completing all retries
	assert.Less(t, mock.RequestCount(), 5, "should have been cancelled before all retries")

	// Should complete within reasonable time (context timeout + some buffer)
	assert.Less(t, elapsed, 3*time.Second, "should respect context cancellation quickly")
}

// =============================================================================
// Edge Case Tests
// =============================================================================

// TestEdgeCases_RequestAndResponseValidation tests various edge cases that could
// cause panics or unexpected behavior in production.
func TestOpenAIProvider_EdgeCases_RequestAndResponseValidation(t *testing.T) {
	t.Run("request_with_empty_messages", func(t *testing.T) {
		mock := newMockServer(t, mockResponse{
			StatusCode: http.StatusOK,
			Body: openai.ChatCompletionResponse{
				Model: "test-model",
				Choices: []openai.ChatCompletionChoice{
					{Message: openai.ChatCompletionMessage{Content: "Empty request handled"}},
				},
			},
		})
		defer mock.Close()

		provider := newTestProvider(t, ProviderConfig{BaseURL: mock.URL()})

		resp, err := provider.ChatCompletion(context.Background(), ChatRequest{
			Messages: []state.Message{}, // Empty messages
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "Empty request handled", resp.Content)
	})

	t.Run("response_with_nil_usage", func(t *testing.T) {
		mock := newMockServer(t, mockResponse{
			StatusCode: http.StatusOK,
			Body: openai.ChatCompletionResponse{
				Model: "test-model",
				Choices: []openai.ChatCompletionChoice{
					{Message: openai.ChatCompletionMessage{Content: "No usage stats"}},
				},
				// Usage field is not set (zero value)
			},
		})
		defer mock.Close()

		provider := newTestProvider(t, ProviderConfig{BaseURL: mock.URL()})

		resp, err := provider.ChatCompletion(context.Background(), ChatRequest{
			Messages: []state.Message{{Role: state.RoleUser, Content: "Test"}},
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "No usage stats", resp.Content)
		// Usage should be zero values, not nil
		assert.Equal(t, 0, resp.Usage.PromptTokens)
		assert.Equal(t, 0, resp.Usage.CompletionTokens)
		assert.Equal(t, 0, resp.Usage.TotalTokens)
	})

	t.Run("very_large_response", func(t *testing.T) {
		// Test handling of very large content
		largeContent := strings.Repeat("This is a very long response. ", 1000)

		mock := newMockServer(t, mockResponse{
			StatusCode: http.StatusOK,
			Body: openai.ChatCompletionResponse{
				Model: "test-model",
				Choices: []openai.ChatCompletionChoice{
					{Message: openai.ChatCompletionMessage{Content: largeContent}},
				},
			},
		})
		defer mock.Close()

		provider := newTestProvider(t, ProviderConfig{BaseURL: mock.URL()})

		resp, err := provider.ChatCompletion(context.Background(), ChatRequest{
			Messages: []state.Message{{Role: state.RoleUser, Content: "Generate long text"}},
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, largeContent, resp.Content)
		assert.Greater(t, len(resp.Content), 10000, "should handle large responses")
	})

	t.Run("unicode_and_special_characters", func(t *testing.T) {
		unicodeContent := "Hello 🌍! Testing émojis, spéçial chars: αβγδε, and 中文字符"

		mock := newMockServer(t, mockResponse{
			StatusCode: http.StatusOK,
			Body: openai.ChatCompletionResponse{
				Model: "test-model",
				Choices: []openai.ChatCompletionChoice{
					{Message: openai.ChatCompletionMessage{Content: unicodeContent}},
				},
			},
		})
		defer mock.Close()

		provider := newTestProvider(t, ProviderConfig{BaseURL: mock.URL()})

		resp, err := provider.ChatCompletion(context.Background(), ChatRequest{
			Messages: []state.Message{{Role: state.RoleUser, Content: "Unicode test"}},
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, unicodeContent, resp.Content)
	})
}

// =============================================================================
// Name and Models Method Tests (minimal but necessary)
// =============================================================================

// TestProviderMetadata tests the simple metadata methods.
// These are kept minimal as they provide little business value to test extensively.
func TestOpenAIProvider_Metadata(t *testing.T) {
	t.Run("lmstudio_provider", func(t *testing.T) {
		provider := newTestProvider(t, ProviderConfig{BaseURL: "http://localhost:1234/v1"})
		
		assert.Equal(t, state.ProviderLMStudio, provider.Name())
		
		models, err := provider.Models(context.Background())
		require.NoError(t, err)
		assert.Empty(t, models, "LMStudio provider should return empty models list")
	})
	
	t.Run("openai_provider", func(t *testing.T) {
		provider := newTestProvider(t, ProviderConfig{BaseURL: "https://api.openai.com/v1"})
		
		assert.Equal(t, state.ProviderOpenAI, provider.Name())
	})
}
