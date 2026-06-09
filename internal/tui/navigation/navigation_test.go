package navigation_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"linux-helper/internal/tui/navigation"
)

type fakeModel struct{}

// Init starts a no-op model.
func (fakeModel) Init() tea.Cmd { return nil }

// Update updates a no-op model.
func (fakeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return fakeModel{}, nil }

// View renders a placeholder view.
func (fakeModel) View() string { return "ok" }

// TestStackPushPop manages the top model.
func TestStackPushPop(t *testing.T) {
	stack := navigation.NewStack(fakeModel{})
	stack.Push(fakeModel{})
	assert.NotNil(t, stack.Top())
	stack.Pop()
	assert.NotNil(t, stack.Top())
}
