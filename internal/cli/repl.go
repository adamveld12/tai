package cli

import (
	"fmt"
	"log"

	"github.com/adamveld12/tai/internal/llm"
	"github.com/adamveld12/tai/internal/state"
	"github.com/adamveld12/tai/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

type ReplHandler struct {
	state.Dispatcher
	llm.Provider
	ui.Stack
	*Config
	*tea.Program
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

	program := tea.NewProgram(stack, tea.WithAltScreen())

	return &ReplHandler{
		Dispatcher: s,
		Provider:   provider,
		Stack:      stack,
		Config:     config,
		Program:    program,
	}
}

func (h *ReplHandler) Execute() error {
	// wire the state change handler to the UI
	h.Dispatcher.OnStateChange(func(a state.Action, as state.AppState, os state.AppState) {
		cmd := h.Stack.Active().OnStateChange(a, as, os)
		h.Program.Send(cmd)
	})

	if _, err := h.Program.Run(); err != nil {
		return fmt.Errorf("😢 failed to start REPL:\n%w", err)
	}

	return nil
}
