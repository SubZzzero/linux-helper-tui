package navigation

import tea "github.com/charmbracelet/bubbletea"

// Stack manages the current screen stack.
type Stack struct {
	screens []tea.Model
}

// NewStack constructs a stack with one root screen.
func NewStack(root tea.Model) *Stack {
	return &Stack{screens: []tea.Model{root}}
}

// Push adds one screen to the top.
func (s *Stack) Push(model tea.Model) {
	s.screens = append(s.screens, model)
}

// Pop removes the top screen when possible.
func (s *Stack) Pop() {
	if len(s.screens) <= 1 {
		return
	}

	s.screens = s.screens[:len(s.screens)-1]
}

// Top returns the current top screen.
func (s *Stack) Top() tea.Model {
	return s.screens[len(s.screens)-1]
}

// ReplaceTop swaps the current screen in place.
func (s *Stack) ReplaceTop(model tea.Model) {
	s.screens[len(s.screens)-1] = model
}

// Root returns the bottom screen in the stack.
func (s *Stack) Root() tea.Model {
	return s.screens[0]
}

// ReplaceRoot swaps the bottom screen in place.
func (s *Stack) ReplaceRoot(model tea.Model) {
	s.screens[0] = model
}
