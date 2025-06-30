package cli

import (
	"fmt"
	"log"

	"github.com/adamveld12/tai/internal/llm"
	"github.com/adamveld12/tai/internal/state"
	"github.com/adamveld12/tai/internal/ui"
)

type ReplHandler struct {
	state.Dispatcher
	llm.Provider
	ui.Stack
	*Config
}

func NewReplHandler(config *Config) *ReplHandler {
	provider, err := llm.NewLMStudioProvider(llm.ProviderConfig{})
	if err != nil {
		log.Fatalf("Failed to initialize LLM provider: %v", err)
	}

	s := state.NewMemoryState(config.SystemPrompt, config.WorkingDirectory, "")
	stack := ui.NewScreenStack(
		ui.NewREPL(s, provider),
	)

	s.OnStateChange(stack.OnStateChange)

	return &ReplHandler{
		Dispatcher: s,
		Provider:   provider,
		Stack:      stack,
		Config:     config,
	}
}

func (h *ReplHandler) Execute() error {
	// wire the state change handler to the UI

	if err := h.Stack.Run(); err != nil {
		return fmt.Errorf("😢 failed to start REPL:\n%w", err)
	}

	return nil
}
