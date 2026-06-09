package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"linux-helper/internal/models"
	uitheme "linux-helper/internal/tui/theme"
)

// Searcher is the screen-local search dependency.
type Searcher interface {
	Search(query string) ([]models.Recipe, error)
}

// SearchModel renders the recipe search screen.
type SearchModel struct {
	service     Searcher
	locale      string
	styles      uitheme.Styles
	input       textinput.Model
	results     []models.Recipe
	selected    int
	pending     *models.Recipe
	title       string
	placeholder string
	emptyText   string
	width       int
	height      int
	err         error
}

// NewSearchModel constructs the initial search screen.
func NewSearchModel(service Searcher, locale string, styles uitheme.Styles, title string, placeholder string, emptyText string) (SearchModel, error) {
	input := textinput.New()
	input.Placeholder = placeholder
	input.Focus()
	input.Prompt = "> "

	results, err := service.Search("")
	if err != nil {
		return SearchModel{}, fmt.Errorf("load initial search results: %w", err)
	}

	return SearchModel{
		service:     service,
		locale:      locale,
		styles:      styles,
		input:       input,
		results:     results,
		title:       title,
		placeholder: placeholder,
		emptyText:   emptyText,
	}, nil
}

// Init starts the screen with no async work.
func (m SearchModel) Init() tea.Cmd {
	return nil
}

// Update handles input and selection changes.
func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
	case tea.KeyMsg:
		switch typed.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "down":
			if m.selected < len(m.results)-1 {
				m.selected++
			}
			return m, nil
		case "enter":
			if len(m.results) > 0 {
				selected := m.results[m.selected]
				m.pending = &selected
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	results, err := m.service.Search(m.input.Value())
	m.err = err
	if err == nil {
		m.results = results
		if m.selected >= len(m.results) {
			m.selected = max(0, len(m.results)-1)
		}
	}

	return m, cmd
}

// ConsumeSelection returns the selected recipe once.
func (m *SearchModel) ConsumeSelection() (models.Recipe, bool) {
	if m.pending == nil {
		return models.Recipe{}, false
	}

	recipe := *m.pending
	m.pending = nil
	return recipe, true
}

// View renders the search UI.
func (m SearchModel) View() string {
	content := []string{m.styles.Title.Render(m.title), m.input.View(), ""}
	if m.err != nil {
		content = append(content, m.styles.Error.Render("Error: "+m.err.Error()))
	} else if len(m.results) == 0 {
		content = append(content, m.styles.Muted.Render(m.emptyText))
	} else {
		content = append(content, m.renderResults())
	}

	return renderFrame(m.styles, m.width, content)
}

// renderResults renders the current result list.
func (m SearchModel) renderResults() string {
	lines := make([]string, 0, len(m.results))
	for index, recipe := range m.results {
		line := "  " + resolveRecipeText(m.locale, recipe.Title) + " [" + string(recipe.Category) + "]"
		if index == m.selected {
			line = m.styles.Selected.Render("> " + resolveRecipeText(m.locale, recipe.Title) + " [" + string(recipe.Category) + "]")
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
