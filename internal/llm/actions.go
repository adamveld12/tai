package llm

import (
	"time"

	"github.com/adamveld12/tai/internal/llm"
	"github.com/adamveld12/tai/internal/state"
)

type ChangeProviderSettings struct {
	Provider state.SupportedProvider
	Model    string
}

func (s ChangeProviderSettings) Execute(state state.AppState) (state.AppState, error) {
	state.Model.Provider = s.Provider
	state.Model.Name = s.Model
	return state, nil
}

func GetProvider(s ChangeProviderSettings) (Provider, error) {
	switch s.Provider {
	case state.ProviderLMStudio:
		if s.Model == "" {
			s.Model = "gemma-3n-e4b-it"
		}
		return llm.NewLMStudioProvider(ProviderConfig{
			DefaultModel: s.Model,
			Timeout:      time.Second * 120,
			MaxRetries:   3,
		})

	}
}
