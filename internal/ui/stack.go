package ui

import tea "github.com/charmbracelet/bubbletea"

// ScreenStack implements a stack of screens. It follows LIFO order.
type ScreenStack struct {
	root        Screen
	screenStack []Screen
}

// Push adds a screen to the top of the stack and returns the new stack size
func (s *ScreenStack) Push(screen Screen) int {
	s.screenStack = append(s.screenStack, screen)
	return len(s.screenStack)
}

// Pop removes and returns the top screen from the stack
// Returns nil if the stack is empty
func (s *ScreenStack) Pop() Screen {
	if len(s.screenStack) == 0 {
		return nil
	}

	// Get the last element
	screen := s.screenStack[len(s.screenStack)-1]

	// Remove it from the stack
	s.screenStack = s.screenStack[:len(s.screenStack)-1]

	return screen
}

// Clear removes all screens from the stack
func (s *ScreenStack) Clear() {
	if len(s.screenStack) > 0 {
		s.screenStack = make([]Screen, 0)
	}
}

// Active returns the top screen from the stack without removing it
// Returns the root screen if the stack is empty
func (s *ScreenStack) Active() Screen {
	if len(s.screenStack) == 0 {
		return s.root
	}
	return s.screenStack[len(s.screenStack)-1]
}

// Init implements tea.Model interface
func (s *ScreenStack) Init() tea.Cmd {
	// Initialize the active screen if it exists
	if active := s.Active(); active != nil {
		return active.Init()
	}
	return nil
}

// Update implements tea.Model interface
func (s *ScreenStack) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Delegate to the active screen if it exists
	if active := s.Active(); active != nil {
		return active.Update(msg)
	}

	return s, nil
}

// View implements tea.Model interface
func (s *ScreenStack) View() string {
	// Render the active screen if it exists
	if active := s.Active(); active != nil {
		return active.View()
	}

	return "💩NOTHIN TO SEE HERE 💩"
}

// NewScreenStack creates a new screen stack
func NewScreenStack(root Screen) *ScreenStack {
	return &ScreenStack{root: root}
}
