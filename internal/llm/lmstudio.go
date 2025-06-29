package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/adamveld12/tai/internal/state"
	"github.com/sashabaranov/go-openai"
)

const ProviderLMStudio SupportedProvider = "lmstudio"

// LMStudioProvider implements the Provider interface for LM Studio
type LMStudioProvider struct {
	client       *openai.Client
	config       ProviderConfig
	defaultModel string
}

// NewLMStudioProvider creates a new LM Studio provider instance
func NewLMStudioProvider(config ProviderConfig) (*LMStudioProvider, error) {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:1234/v1"
	}

	if config.APIKey == "" {
		// LM Studio doesn't require an API key, but the OpenAI client expects one
		config.APIKey = "lm-studio"
	}

	if config.DefaultModel == "" {
		config.DefaultModel = "gemma-3n-e4b-it"
	}

	if config.Timeout == 0 {
		config.Timeout = 300 * time.Second
	}

	clientConfig := openai.DefaultConfig(config.APIKey)
	clientConfig.BaseURL = config.BaseURL
	client := openai.NewClientWithConfig(clientConfig)

	return &LMStudioProvider{
		client:       client,
		config:       config,
		defaultModel: config.DefaultModel,
	}, nil
}

// Name returns the provider name
func (p *LMStudioProvider) Name() SupportedProvider {
	return ProviderLMStudio
}

// ChatCompletion sends a chat completion request and returns the response
func (p *LMStudioProvider) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Convert our ChatRequest to OpenAI format
	openAIReq := p.convertToOpenAIRequest(req, false)

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	defer cancel()

	startTime := time.Now()

	var resp openai.ChatCompletionResponse
	err := p.retryRequest(ctx, func() error {
		var err error
		resp, err = p.client.CreateChatCompletion(ctx, openAIReq)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("chat completion failed: %w", err)
	}

	duration := time.Since(startTime)

	// Convert the response back to our format
	return p.convertFromOpenAIResponse(resp, duration), nil
}

// StreamChatCompletion sends a streaming chat completion request
func (p *LMStudioProvider) StreamChatCompletion(ctx context.Context, req ChatRequest) (<-chan ChatStreamChunk, error) {
	// Convert our ChatRequest to OpenAI format
	openAIReq := p.convertToOpenAIRequest(req, true)

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, p.config.Timeout)

	// Create the stream
	stream, err := p.client.CreateChatCompletionStream(ctx, openAIReq)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("stream creation failed: %w", err)
	}

	// Create channel for chunks
	chunkChan := make(chan ChatStreamChunk)

	// Start goroutine to process stream
	go func() {
		defer close(chunkChan)
		defer cancel()
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				// Send final chunk
				chunkChan <- ChatStreamChunk{Done: true}
				return
			}

			if err != nil {
				chunkChan <- ChatStreamChunk{Error: fmt.Errorf("stream error: %w", err), Done: true}
				return
			}

			// Convert response to our chunk format
			if len(response.Choices) > 0 {
				var usage TokenUsage
				if response.Usage != nil {
					usage = TokenUsage{
						PromptTokens:     response.Usage.PromptTokens,
						CompletionTokens: response.Usage.CompletionTokens,
						TotalTokens:      response.Usage.TotalTokens,
					}
				}

				chunk := ChatStreamChunk{
					Usage: usage,
					Model: response.Model,
					Delta: response.Choices[0].Delta.Content,
					Done:  false,
				}

				// Handle tool calls if present
				if len(response.Choices[0].Delta.ToolCalls) > 0 {
					chunk.ToolCalls = p.convertToolCallsFromOpenAI(response.Choices[0].Delta.ToolCalls)
				}

				select {
				case chunkChan <- chunk:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return chunkChan, nil
}

// We don't support listing models for LM Studio as it typically runs local models
// Empty list is returned to indicate no specific models are available
func (p *LMStudioProvider) Models(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

// convertToOpenAIRequest converts our ChatRequest to OpenAI format
func (p *LMStudioProvider) convertToOpenAIRequest(req ChatRequest, stream bool) openai.ChatCompletionRequest {
	model := req.Model
	if model == "" {
		model = p.defaultModel
	}

	openAIReq := openai.ChatCompletionRequest{
		Model:    model,
		Messages: make([]openai.ChatCompletionMessage, 0, len(req.Messages)),
		Stream:   stream,
	}

	// Set temperature if provided
	if req.Temperature > 0 {
		openAIReq.Temperature = float32(req.Temperature)
	}

	// Set max tokens if provided
	if req.MaxTokens > 0 {
		openAIReq.MaxTokens = req.MaxTokens
	}

	// Convert messages
	for _, msg := range req.Messages {
		openAIMsg := openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
			Name:    "",
		}

		// // Handle tool calls
		if len(req.Tools) > 0 {
			openAIMsg.ToolCalls = p.convertToolCallsToOpenAI(msg.ToolCalls)
		}

		openAIReq.Messages = append(openAIReq.Messages, openAIMsg)
	}

	// Handle system prompt
	if req.SystemPrompt != "" {
		// Prepend system message if not already present
		if len(openAIReq.Messages) == 0 || openAIReq.Messages[0].Role != string(state.RoleSystem) {
			systemMsg := openai.ChatCompletionMessage{
				Role:    string(state.RoleSystem),
				Content: req.SystemPrompt,
			}
			openAIReq.Messages = append([]openai.ChatCompletionMessage{systemMsg}, openAIReq.Messages...)
		}
	}

	// Convert tools
	if len(req.Tools) > 0 {
		openAIReq.Tools = make([]openai.Tool, 0, len(req.Tools))
		for _, tool := range req.Tools {
			openAITool := openai.Tool{
				Type: openai.ToolType(tool.Type),
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
			openAIReq.Tools = append(openAIReq.Tools, openAITool)
		}
	}

	// Set tool choice
	if req.ToolChoice != "" {
		openAIReq.ToolChoice = req.ToolChoice
	}

	return openAIReq
}

// convertFromOpenAIResponse converts OpenAI response to our format
func (p *LMStudioProvider) convertFromOpenAIResponse(resp openai.ChatCompletionResponse, duration time.Duration) *ChatResponse {
	response := &ChatResponse{
		Model:     resp.Model,
		CreatedAt: time.Unix(resp.Created, 0),
		Duration:  duration,
		Usage: TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	// Extract content and finish reason from first choice
	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		response.Content = choice.Message.Content
		response.FinishReason = string(choice.FinishReason)

		// Convert tool calls if present
		if len(choice.Message.ToolCalls) > 0 {
			response.ToolCalls = p.convertToolCallsFromOpenAI(choice.Message.ToolCalls)
		}
	}

	return response
}

// convertToolCallsToOpenAI converts our tool calls to OpenAI format
func (p *LMStudioProvider) convertToolCallsToOpenAI(toolCalls []state.ToolCall) []openai.ToolCall {
	openAIToolCalls := make([]openai.ToolCall, 0, len(toolCalls))
	for _, tc := range toolCalls {
		openAIToolCalls = append(openAIToolCalls, openai.ToolCall{
			ID:   tc.ID,
			Type: openai.ToolType(tc.Type),
			Function: openai.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}
	return openAIToolCalls
}

// convertToolCallsFromOpenAI converts OpenAI tool calls to our format
func (p *LMStudioProvider) convertToolCallsFromOpenAI(openAIToolCalls []openai.ToolCall) []state.ToolCall {
	toolCalls := make([]state.ToolCall, 0, len(openAIToolCalls))
	for _, tc := range openAIToolCalls {
		toolCalls = append(toolCalls, state.ToolCall{
			ID:   tc.ID,
			Type: string(tc.Type),
			Function: state.ToolCallFunction{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}
	return toolCalls
}

// Retry logic for failed requests
func (p *LMStudioProvider) retryRequest(ctx context.Context, fn func() error) error {
	maxRetries := p.config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := fn(); err != nil {
			lastErr = err

			// Check if context is cancelled
			if ctx.Err() != nil {
				return ctx.Err()
			}

			// Don't retry on certain errors
			if strings.Contains(err.Error(), "invalid_api_key") ||
				strings.Contains(err.Error(), "model_not_found") {
				return err
			}

			// Exponential backoff
			if i < maxRetries-1 {
				backoff := time.Duration(1<<uint(i)) * time.Second
				select {
				case <-time.After(backoff):
					// Continue to next retry
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}
