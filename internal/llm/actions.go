package llm

import (
	"time"

	"github.com/adamveld12/tai/internal/state"
)

type ChangeProviderSettingsAction struct {
	Provider state.SupportedProvider
	Model    string
}

func (s ChangeProviderSettingsAction) Execute(state state.AppState) (state.AppState, error) {
	if s.Provider != "" {
		state.Model.Provider = s.Provider
	}

	if s.Model != "" {
		state.Model.Name = s.Model
	}
	return state, nil
}

func GetProvider(d state.Dispatcher, p state.SupportedProvider, model string) (pr Provider, err error) {
	pc := ProviderConfig{
		DefaultModel: model,
		Timeout:      time.Second * 120,
		MaxRetries:   2,
	}

	switch p {
	case state.ProviderOpenAI:
		pr, err = NewOpenAIProvider(pc)
	case state.ProviderLMStudio:
		if pc.DefaultModel == "" {
			pc.DefaultModel = "gemma-3n-e4b-it"
		}
		pr, err = NewLMStudioProvider(pc)
	default:
		return
	}

	if err != nil {
		return
	}

	d.Dispatch(ChangeProviderSettingsAction{
		Provider: p,
		Model:    pr.Model(),
	})
	return
}
