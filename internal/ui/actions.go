package ui

import (
	"time"

	"github.com/adamveld12/tai/internal/state"
)

func ClearMessages(d state.Dispatcher) {
	d.Dispatch(ClearMessagesAction{})
}

type ClearMessagesAction struct{}

type SwitchThemeAction struct {
	Name string
}

func (a ClearMessagesAction) Execute(s state.AppState) (state.AppState, error) {
	s.Context.Messages = []state.Message{}
	s.Model.Busy = false
	s.Context.Updated = time.Now()
	return s, nil
}

type ChangeProviderAction struct {
	Provider string
	Name     string
}

func (a ChangeProviderAction) Execute(s state.AppState) (state.AppState, error) {
	s.Model.Provider = state.SupportedProvider(a.Provider)
	s.Model.Name = a.Name
	s.Model.Busy = true
	return s, nil

}
