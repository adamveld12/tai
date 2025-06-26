package ui

import (
	"github.com/adamveld12/tai/internal/state"
	tea "github.com/charmbracelet/bubbletea"
)

type Screen interface {
	OnStateChange(action state.Action, newState state.AppState, oldState state.AppState) tea.Msg
	tea.Model
}

// Stack defines the interface for a screen stack
type Stack interface {
	tea.Model
	Active() Screen
	Push(Screen) int
	Pop() Screen
	Clear()
}
