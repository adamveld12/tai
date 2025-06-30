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
	// Determine provider based on config
	var providerType state.SupportedProvider
	switch config.Provider {
	case "lmstudio":
		providerType = state.ProviderLMStudio // default
	case "openai":
		providerType = state.ProviderOpenAI
	default:
		log.Fatal("💩 unknown provider, must be 'openai' or 'lmstudio'")
	}

	s := state.NewMemoryState(config.SystemPrompt, config.WorkingDirectory, "")
	settings := llm.ChangeProviderSettingsAction{
		Provider: providerType,
		Model:    "", // Use default model
	}
	s.Dispatch(settings)

	stack := ui.NewScreenStack(
		ui.NewREPL(s),
	)

	return &ReplHandler{
		Dispatcher: s,
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
