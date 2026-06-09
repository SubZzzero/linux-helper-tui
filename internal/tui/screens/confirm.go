package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"linux-helper/internal/models"
	uitheme "linux-helper/internal/tui/theme"
)

// ConfirmModel asks for explicit approval before risky execution.
type ConfirmModel struct {
	recipe      models.Recipe
	locale      string
	styles      uitheme.Styles
	preview     string
	titleText   string
	approveText string
	backText    string
	pendingBack bool
	pendingRun  bool
	width       int
}

// NewConfirmModel constructs one confirmation screen.
func NewConfirmModel(recipe models.Recipe, locale string, styles uitheme.Styles, preview string, titleText string, approveText string, backText string) ConfirmModel {
	return ConfirmModel{
		recipe:      recipe,
		locale:      locale,
		styles:      styles,
		preview:     preview,
		titleText:   titleText,
		approveText: approveText,
		backText:    backText,
	}
}

// Init starts the confirmation screen with no async work.
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update handles explicit confirmation or cancellation.
func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
	case tea.KeyMsg:
		switch typed.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter", "y":
			m.pendingRun = true
			return m, nil
		case "esc", "n", "backspace":
			m.pendingBack = true
			return m, nil
		}
	}

	return m, nil
}

// ConsumeBack reports whether the confirmation screen requested a pop.
func (m *ConfirmModel) ConsumeBack() bool {
	back := m.pendingBack
	m.pendingBack = false
	return back
}

// ConsumeConfirm reports whether the confirmation screen approved execution.
func (m *ConfirmModel) ConsumeConfirm() bool {
	confirmed := m.pendingRun
	m.pendingRun = false
	return confirmed
}

// View renders the confirmation prompt and command preview.
func (m ConfirmModel) View() string {
	lines := []string{
		m.styles.Title.Render(m.titleText),
		"",
		resolveRecipeText(m.locale, m.recipe.Title),
		fmt.Sprintf("Risk: %s", m.recipe.Risk),
		"",
		m.styles.Accent.Render("Command:"),
		m.preview,
		"",
		m.styles.Accent.Render(m.approveText),
		m.styles.Muted.Render(m.backText),
	}

	return renderFrame(m.styles, m.width, lines)
}
