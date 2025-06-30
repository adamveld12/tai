package state

import (
	"time"
)

type Mode string

const (
	PlanMode    Mode = "plan"
	ExecuteMode Mode = "execute"
	YoloMode    Mode = "yolo"
)

// Role represents the role of a message sender
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

// ToolCall represents a call to a tool made by the LLM
type ToolCall struct {
	// Unique ID for this tool call
	ID string `json:"id"`

	// Type of tool call (currently only "function")
	Type string `json:"type"`

	// Function call details
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction represents the function call details
type ToolCallFunction struct {
	// Name of the function to call
	Name string `json:"name"`

	// Arguments to pass to the function (JSON string)
	Arguments string `json:"arguments"`
}

// Message represents a single message in a conversation
type Message struct {
	Role      Role       `json:"role"`
	Content   string     `json:"content"`
	Usage     TokenUsage `json:"usage"`
	ToolCalls []ToolCall `json:"toolCalls"`
	Timestamp time.Time  `json:"timestamp"`
}

type TokenUsage struct {
	Prompt     int `json:"prompt"`
	Completion int `json:"completion"`
	Total      int `json:"total"`
}

type AppState struct {
	Permissions Permissions `json:"permissions"`
	Context     Context     `json:"context"`
	Model       Model       `json:"model"`
	Status      struct {
		Error error `json:"error,omitempty"`
	}
}

type Permissions struct {
	Allow []string `json:"allow"`
	Deny  []string `json:"deny"`
}

type Context struct {
	Mode             Mode      `json:"mode"`
	SystemPrompt     string    `json:"systemPrompt"`
	SessionID        string    `json:"sessionId"`
	Messages         []Message `json:"messages"`
	PromptTokens     int       `json:"promptTokens"`
	CompletionTokens int       `json:"completionTokens"`
	Created          time.Time `json:"created"`
	Updated          time.Time `json:"updated"`
	WorkingDirectory string    `json:"workingDirectory"`
}

type SupportedProvider string

const (
	ModelNotAvailableError                   = "model not available"
	ProviderOpenAI         SupportedProvider = "openai"
	ProviderLMStudio       SupportedProvider = "lmstudio"
)

type Model struct {
	Provider SupportedProvider `json:"provider"`
	Name     string            `json:"name"`
	Busy     bool              `json:"busy"`
}

type ActionID string

type Action interface {
	Execute(state AppState) (AppState, error)
}

type OnStateChangeHandler func(Action, AppState, AppState)
type Dispatcher interface {
	GetState() AppState
	OnStateChange(OnStateChangeHandler)
	Dispatch(Action)
}
