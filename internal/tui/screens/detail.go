package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"linux-helper/internal/models"
	uitheme "linux-helper/internal/tui/theme"
)

// DetailModel renders one recipe detail screen.
type DetailModel struct {
	recipe   models.Recipe
	locale   string
	styles   uitheme.Styles
	runText  string
	backText string
	start    bool
	back     bool
	width    int
}

// NewDetailModel constructs a detail screen.
func NewDetailModel(recipe models.Recipe, locale string, styles uitheme.Styles, runText string, backText string) DetailModel {
	return DetailModel{recipe: recipe, locale: locale, styles: styles, runText: runText, backText: backText}
}

// Init starts the detail screen with no async work.
func (m DetailModel) Init() tea.Cmd {
	return nil
}

// Update handles back navigation.
func (m DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
	case tea.KeyMsg:
		switch typed.String() {
		case "esc", "backspace":
			m.back = true
		case "enter":
			m.start = true
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

// ConsumeExecute reports whether the screen requested form entry.
func (m *DetailModel) ConsumeExecute() bool {
	start := m.start
	m.start = false
	return start
}

// Recipe returns the selected recipe.
func (m DetailModel) Recipe() models.Recipe {
	return m.recipe
}

// ConsumeBack reports whether the screen requested a pop.
func (m *DetailModel) ConsumeBack() bool {
	back := m.back
	m.back = false
	return back
}

// View renders the selected recipe details.
func (m DetailModel) View() string {
	lines := []string{
		m.styles.Title.Render(resolveRecipeText(m.locale, m.recipe.Title)),
		"",
		resolveRecipeText(m.locale, m.recipe.Description),
		"",
		"Category: " + m.recipe.Category.DisplayName(),
		"Risk: " + string(m.recipe.Risk),
		"Execution: " + string(m.recipe.Execution),
		"",
		m.styles.Accent.Render("Args:"),
		strings.Join(m.recipe.Args, " "),
		"",
		m.styles.Accent.Render(m.runText),
		"",
		m.styles.Muted.Render(m.backText),
	}

	return renderFrame(m.styles, m.width, lines)
}
