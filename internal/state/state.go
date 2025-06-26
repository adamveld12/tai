package state

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type MemoryState struct {
	state     AppState
	mu        sync.RWMutex
	listeners []OnStateChangeHandler
}

// NewMemoryState creates a new MemoryState instance
func NewMemoryState(systemPrompt, workingDirectory, sessionName string) *MemoryState {
	now := time.Now()
	if systemPrompt == "" {
		systemPrompt = "You are an AI assistant that autonomously writes code and helps the user with programming tasks."
	}

	if workingDirectory == "" {
		var err error
		workingDirectory, err = os.Getwd()
		if err != nil {
			workingDirectory = "."
		}
	}

	if sessionName == "" {
		sessionName = now.Format("session-20060102150405")
	}

	state := AppState{
		Context: Context{
			Created:          now,
			Updated:          now,
			Mode:             PlanMode,
			SystemPrompt:     systemPrompt,
			WorkingDirectory: workingDirectory,
			SessionID:        sessionName,
		},
	}

	return &MemoryState{
		state:     state,
		listeners: make([]OnStateChangeHandler, 0),
		mu:        sync.RWMutex{},
	}
}

// Returns a copy of the current state to prevent external mutations
func (m *MemoryState) GetState() AppState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

func (m *MemoryState) OnStateChange(listener OnStateChangeHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}

// dispatch implements Dispatcher interface
func (m *MemoryState) Dispatch(action Action) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldState := m.state
	// Execute the action to get the new state
	newState, err := action.Execute(m.state)
	if err != nil {
		// TODO: Handle error - for now, we'll just return without updating state
		// In a production app, you might want to log this or handle it differently
		panic(fmt.Errorf("💩 failed to execute action %v: %w", action, err))
	}

	newState.Context.Updated = time.Now()
	m.state = newState

	// Notify all listeners
	for _, listener := range m.listeners {
		go listener(action, newState, oldState)
	}
}
