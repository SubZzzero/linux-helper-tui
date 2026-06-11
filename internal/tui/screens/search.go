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
	service         Searcher
	locale          string
	styles          uitheme.Styles
	favorites       map[string]struct{}
	recent          []string
	input           textinput.Model
	results         []models.Recipe
	selected        int
	pending         *models.Recipe
	pendingFavorite *models.Recipe
	title           string
	placeholder     string
	emptyText       string
	recentTitle     string
	recentEmpty     string
	helpText        string
	width           int
	height          int
	err             error
}

// NewSearchModel constructs the initial search screen.
func NewSearchModel(service Searcher, locale string, styles uitheme.Styles, favorites []string, recent []string, title string, placeholder string, emptyText string, recentTitle string, recentEmpty string, helpText string) (SearchModel, error) {
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
		favorites:   favoriteSet(favorites),
		recent:      append([]string(nil), recent...),
		input:       input,
		results:     orderResults(results, favoriteSet(favorites)),
		title:       title,
		placeholder: placeholder,
		emptyText:   emptyText,
		recentTitle: recentTitle,
		recentEmpty: recentEmpty,
		helpText:    helpText,
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
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k", "ctrl+p":
			if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "down", "j", "ctrl+n":
			if m.selected < len(m.results)-1 {
				m.selected++
			}
			return m, nil
		case "home", "g":
			if len(m.results) > 0 {
				m.selected = 0
			}
			return m, nil
		case "end", "G":
			if len(m.results) > 0 {
				m.selected = len(m.results) - 1
			}
			return m, nil
		case "f":
			if len(m.results) > 0 {
				selected := m.results[m.selected]
				m.pendingFavorite = &selected
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
		m.results = orderResults(results, m.favorites)
		if m.selected >= len(m.results) {
			m.selected = max(0, len(m.results)-1)
		}
	}

	return m, cmd
}

// ConsumeToggleFavorite returns the selected recipe for a favorite toggle once.
func (m *SearchModel) ConsumeToggleFavorite() (models.Recipe, bool) {
	if m.pendingFavorite == nil {
		return models.Recipe{}, false
	}

	recipe := *m.pendingFavorite
	m.pendingFavorite = nil
	return recipe, true
}

// SetFavorites updates the favorite state used for rendering and ordering.
func (m *SearchModel) SetFavorites(favorites []string) {
	m.favorites = favoriteSet(favorites)
	m.results = orderResults(m.results, m.favorites)
	if m.selected >= len(m.results) {
		m.selected = max(0, len(m.results)-1)
	}
}

// SetRecent updates the recent command list.
func (m *SearchModel) SetRecent(recent []string) {
	m.recent = append([]string(nil), recent...)
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

	content = append(content, "", m.styles.Accent.Render(m.recentTitle))
	if len(m.recent) == 0 {
		content = append(content, m.styles.Muted.Render(m.recentEmpty))
	} else {
		content = append(content, m.renderRecent())
	}
	content = append(content, "", m.styles.Muted.Render(m.helpText))

	return renderFrame(m.styles, m.width, content)
}

// renderResults renders the current result list.
func (m SearchModel) renderResults() string {
	lines := make([]string, 0, len(m.results))
	for index, recipe := range m.results {
		line := "  " + favoriteMarker(m.favorites, recipe.ID) + " " + resolveRecipeText(m.locale, recipe.Title) + " [" + string(recipe.Category) + "]"
		if index == m.selected {
			line = m.styles.Selected.Render("> " + favoriteMarker(m.favorites, recipe.ID) + " " + resolveRecipeText(m.locale, recipe.Title) + " [" + string(recipe.Category) + "]")
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m SearchModel) renderRecent() string {
	lines := make([]string, 0, min(maxRecentVisible, len(m.recent)))
	for index, command := range m.recent {
		if index >= maxRecentVisible {
			break
		}

		lines = append(lines, "- "+command)
	}

	return strings.Join(lines, "\n")
}

func favoriteSet(favorites []string) map[string]struct{} {
	set := make(map[string]struct{}, len(favorites))
	for _, favoriteID := range favorites {
		set[favoriteID] = struct{}{}
	}

	return set
}

func orderResults(results []models.Recipe, favorites map[string]struct{}) []models.Recipe {
	ordered := make([]models.Recipe, 0, len(results))
	for _, recipe := range results {
		if _, ok := favorites[recipe.ID]; ok {
			ordered = append(ordered, recipe)
		}
	}
	for _, recipe := range results {
		if _, ok := favorites[recipe.ID]; !ok {
			ordered = append(ordered, recipe)
		}
	}

	return ordered
}

func favoriteMarker(favorites map[string]struct{}, recipeID string) string {
	if _, ok := favorites[recipeID]; ok {
		return "[*]"
	}

	return "[ ]"
}

const maxRecentVisible = 5
