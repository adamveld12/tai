package state

import (
	"sync"
	"testing"
)

// mockAction is a test implementation of the Action interface
type mockAction struct {
	name     string
	execFunc func(AppState) (AppState, error)
}

func (m *mockAction) Execute(state AppState) (AppState, error) {
	if m.execFunc != nil {
		return m.execFunc(state)
	}
	return state, nil
}

// Test that Action interface can be implemented and enforces immutability
func TestAction_Interface(t *testing.T) {
	action := &mockAction{
		name: "test-action",
		execFunc: func(state AppState) (AppState, error) {
			newState := state
			newState.Context.Mode = ExecuteMode
			return newState, nil
		},
	}

	// Test that mockAction properly implements Action interface
	var _ Action = action

	// Test Execute method
	initialState := AppState{
		Context: Context{Mode: PlanMode},
	}

	newState, err := action.Execute(initialState)
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
	if newState.Context.Mode != ExecuteMode {
		t.Errorf("Execute() Mode = %q, want %q", newState.Context.Mode, ExecuteMode)
	}

	// Ensure original state wasn't modified
	if initialState.Context.Mode != PlanMode {
		t.Error("Execute() should not modify the original state")
	}
}

// Test that Dispatcher interface can be implemented
func TestDispatcher_Interface(t *testing.T) {
	ms := NewMemoryState("test", "/test", "test")

	// Test that MemoryState implements Dispatcher interface
	var dispatcher Dispatcher = ms

	// Test GetState
	state := dispatcher.GetState()
	if state.Context.SystemPrompt != "test" {
		t.Errorf("GetState() SystemPrompt = %q, want %q", state.Context.SystemPrompt, "test")
	}

	// Test OnStateChange with proper synchronization
	var called bool
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(1)
	dispatcher.OnStateChange(func(action Action, oldState, newState AppState) {
		mu.Lock()
		called = true
		mu.Unlock()
		wg.Done()
	})

	// Test Dispatch
	action := &mockAction{
		name: "test-dispatch",
		execFunc: func(state AppState) (AppState, error) {
			return state, nil
		},
	}

	dispatcher.Dispatch(action)

	// Wait for listener to execute
	wg.Wait()

	mu.Lock()
	wasCalled := called
	mu.Unlock()

	if !wasCalled {
		t.Error("OnStateChange listener was not called")
	}
}
