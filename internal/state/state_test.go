package state

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// mockAction is defined in interfaces_test.go and shared between test files

func TestNewMemoryState_DefaultValues(t *testing.T) {
	// Test default system prompt behavior
	ms := NewMemoryState("", "/test", "test")
	state := ms.GetState()

	if state.Context.SystemPrompt != "You are an AI assistant that autonomously writes code and helps the user with programming tasks." {
		t.Errorf("Default SystemPrompt = %q, want %q", state.Context.SystemPrompt, "You are an AI assistant that autonomously writes code and helps the user with programming tasks.")
	}

	// Test that initial mode is PlanMode
	if state.Context.Mode != PlanMode {
		t.Errorf("Default Mode = %q, want %q", state.Context.Mode, PlanMode)
	}

	// Test timestamps are set and equal
	if state.Context.Created.IsZero() || state.Context.Updated.IsZero() {
		t.Error("Timestamps should not be zero")
	}
	if !state.Context.Created.Equal(state.Context.Updated) {
		t.Error("Created and Updated should be equal on initialization")
	}
}

func TestMemoryState_GetState(t *testing.T) {
	ms := NewMemoryState("Test prompt", "/test", "test-session")

	// Get state multiple times to ensure it returns a copy
	state1 := ms.GetState()
	state2 := ms.GetState()

	// Verify both states have the same values
	if state1.Context.SystemPrompt != state2.Context.SystemPrompt {
		t.Error("GetState should return consistent values")
	}

	// Modify state1 and ensure state2 is not affected (proving it's a copy)
	state1.Context.SystemPrompt = "Modified prompt"

	// Get a fresh state and verify it wasn't modified
	state3 := ms.GetState()
	if state3.Context.SystemPrompt == "Modified prompt" {
		t.Error("GetState should return a copy, not a reference")
	}
}

func TestMemoryState_OnStateChange(t *testing.T) {
	ms := NewMemoryState("Test", "/test", "test")

	// Track listener calls
	var listenerCalls []string
	var mu sync.Mutex

	// Add multiple listeners
	ms.OnStateChange(func(action Action, oldState, newState AppState) {
		mu.Lock()
		defer mu.Unlock()
		listenerCalls = append(listenerCalls, "listener1-called")
	})

	ms.OnStateChange(func(action Action, oldState, newState AppState) {
		mu.Lock()
		defer mu.Unlock()
		listenerCalls = append(listenerCalls, "listener2-called")
	})

	// Dispatch an action to trigger listeners
	action := &mockAction{
		name: "test-action",
		execFunc: func(state AppState) (AppState, error) {
			return state, nil
		},
	}

	ms.Dispatch(action)

	// Wait for listeners to execute
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Verify both listeners were called
	if len(listenerCalls) != 2 {
		t.Errorf("Expected 2 listener calls, got %d", len(listenerCalls))
	}
}

func TestMemoryState_Dispatch(t *testing.T) {
	tests := []struct {
		name            string
		action          Action
		expectListeners bool
		listenerCount   int
	}{
		{
			name: "successful action",
			action: &mockAction{
				name: "test-action",
				execFunc: func(state AppState) (AppState, error) {
					newState := state
					newState.Context.SystemPrompt = "Modified by action"
					return newState, nil
				},
			},
			expectListeners: true,
			listenerCount:   2,
		},
		{
			name: "action returns error",
			action: &mockAction{
				name: "error-action",
				execFunc: func(state AppState) (AppState, error) {
					return state, errors.New("action error")
				},
			},
			expectListeners: false,
			listenerCount:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMemoryState("Initial prompt", "/test", "test")

			// Track listener calls
			listenerCalls := make(chan string, tt.listenerCount)
			var wg sync.WaitGroup

			// Add listeners
			for i := 0; i < tt.listenerCount; i++ {
				listenerID := i
				if tt.expectListeners {
					wg.Add(1)
				}
				ms.OnStateChange(func(action Action, oldState, newState AppState) {
					listenerCalls <- fmt.Sprintf("listener%d-called", listenerID)
					if tt.expectListeners {
						wg.Done()
					}
				})
			}

			// Dispatch action - handle panics for error actions
			if tt.name == "action returns error" {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic for error action, but no panic occurred")
					}
				}()
				ms.Dispatch(tt.action)
				// If we reach here, the panic didn't happen (test will fail via defer)
				return
			}

			ms.Dispatch(tt.action)

			if tt.expectListeners {
				// Wait for all listeners to be called
				done := make(chan bool)
				go func() {
					wg.Wait()
					close(done)
				}()

				select {
				case <-done:
					// Success
				case <-time.After(100 * time.Millisecond):
					t.Error("Listeners were not called within timeout")
				}

				// Verify state was updated
				if tt.name == "successful action" {
					state := ms.GetState()
					if state.Context.SystemPrompt != "Modified by action" {
						t.Errorf("State not updated: SystemPrompt = %q, want %q",
							state.Context.SystemPrompt, "Modified by action")
					}
				}
			} else {
				// For error case, ensure listeners weren't called
				select {
				case call := <-listenerCalls:
					t.Errorf("Listener was called when it shouldn't have been: %s", call)
				case <-time.After(50 * time.Millisecond):
					// Success - no listeners called
				}

				// Verify state was not updated
				state := ms.GetState()
				if state.Context.SystemPrompt != "Initial prompt" {
					t.Error("State should not be updated when action returns error")
				}
			}
		})
	}
}

func TestMemoryState_ConcurrentDispatch(t *testing.T) {
	ms := NewMemoryState("Initial", "/test", "test")

	// Track all state changes
	var stateChanges []string
	var mu sync.Mutex

	ms.OnStateChange(func(action Action, oldState, newState AppState) {
		mu.Lock()
		defer mu.Unlock()
		// Since we can't get the action name directly, we'll just track that a change occurred
		stateChanges = append(stateChanges, "state-changed")
	})

	// Dispatch multiple actions concurrently
	var wg sync.WaitGroup
	actionCount := 10

	for i := 0; i < actionCount; i++ {
		wg.Add(1)
		actionName := fmt.Sprintf("action-%d", i)
		go func(name string) {
			defer wg.Done()
			ms.Dispatch(&mockAction{
				name: name,
				execFunc: func(state AppState) (AppState, error) {
					return state, nil
				},
			})
		}(actionName)
	}

	// Wait for all dispatches to complete
	wg.Wait()

	// Give listeners time to execute
	time.Sleep(50 * time.Millisecond)

	// Verify all actions were processed
	mu.Lock()
	defer mu.Unlock()
	if len(stateChanges) != actionCount {
		t.Errorf("Expected %d state changes, got %d", actionCount, len(stateChanges))
	}
}
