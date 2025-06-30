package llm

import (
	"context"
	"time"

	"github.com/adamveld12/tai/internal/state"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// Name returns the provider name (e.g., "openai", "claude", "ollama")
	Name() state.SupportedProvider

	// ChatCompletion sends a chat completion request and returns the response
	ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// StreamChatCompletion sends a streaming chat completion request
	StreamChatCompletion(ctx context.Context, req ChatRequest) (<-chan ChatStreamChunk, error)

	// Models returns the list of available models for this provider
	Models(ctx context.Context) ([]string, error)

	Model() string
}

// ChatRequest represents a request to the language model
type ChatRequest struct {
	// Messages in the conversation
	Messages []state.Message `json:"messages"`

	// Model to use for the request
	Model string `json:"model"`

	// Maximum tokens to generate
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature for response randomness (0.0 to 1.0)
	Temperature float64 `json:"temperature,omitempty"`

	// Whether to stream the response
	Stream bool `json:"stream,omitempty"`

	// System prompt override
	SystemPrompt string `json:"system_prompt,omitempty"`

	// Tools available for function calling
	Tools []Tool `json:"tools,omitempty"`

	// Whether the model can call tools
	ToolChoice string `json:"tool_choice,omitempty"`
}

// ChatResponse represents a response from the language model
type ChatResponse struct {
	// Generated content
	Content string `json:"content"`

	// Tool calls made by the model
	ToolCalls []state.ToolCall `json:"tool_calls,omitempty"`

	// Usage statistics
	Usage TokenUsage `json:"usage"`

	// Response metadata
	Model     string        `json:"model"`
	CreatedAt time.Time     `json:"created_at"`
	Duration  time.Duration `json:"duration"`

	// Finish reason
	FinishReason string `json:"finish_reason"`
}

// ChatStreamChunk represents a chunk in a streaming response
type ChatStreamChunk struct {
	Model string `json:"model"`

	// Delta content for this chunk
	Delta string `json:"delta"`

	// Tool calls in this chunk
	ToolCalls []state.ToolCall `json:"tool_calls,omitempty"`

	// Usage statistics
	Usage TokenUsage `json:"usage"`

	// Whether this is the final chunk
	Done bool `json:"done"`

	// Error if something went wrong
	Error error `json:"error,omitempty"`
}

// Tool represents a tool that can be called by the LLM
type Tool struct {
	// Type of the tool (currently only "function")
	Type string `json:"type"`

	// Function definition
	Function ToolFunction `json:"function"`
}

// ToolFunction defines a function that can be called
type ToolFunction struct {
	// Name of the function
	Name string `json:"name"`

	// Description of what the function does
	Description string `json:"description"`

	// JSON schema for the function parameters
	Parameters map[string]interface{} `json:"parameters"`
}

// TokenUsage represents token usage statistics
type TokenUsage struct {
	// Tokens in the prompt
	PromptTokens int `json:"prompt_tokens"`

	// Tokens in the completion
	CompletionTokens int `json:"completion_tokens"`

	// Total tokens used
	TotalTokens int `json:"total_tokens"`
}

// ProviderLimits represents rate limits and usage information
type ProviderLimits struct {
	// Requests per minute
	RequestsPerMinute int `json:"requests_per_minute"`

	// Tokens per minute
	TokensPerMinute int `json:"tokens_per_minute"`

	// Current usage
	CurrentRequests int `json:"current_requests"`
	CurrentTokens   int `json:"current_tokens"`

	// Reset time for rate limits
	ResetTime time.Time `json:"reset_time"`
}

// ProviderConfig represents configuration for a provider
type ProviderConfig struct {
	// API key or token
	APIKey string `json:"api_key"`

	// Base URL for the API
	BaseURL string `json:"base_url,omitempty"`

	// Default model to use
	DefaultModel string `json:"default_model"`

	// Timeout for requests
	Timeout time.Duration `json:"timeout"`

	// Maximum retries on failure
	MaxRetries int `json:"max_retries"`
}
