package screens_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"linux-helper/internal/models"
	"linux-helper/internal/tui/screens"
	uitheme "linux-helper/internal/tui/theme"
)

type fakeSearcher struct{}

func testStyles() uitheme.Styles {
	return uitheme.NewStyles(uitheme.Definition{Name: "test", BorderColor: "63", AccentColor: "213"})
}

// Search returns static recipes from multiple categories.
func (fakeSearcher) Search(query string) ([]models.Recipe, error) {
	return []models.Recipe{{
		ID:          "find-file",
		Category:    models.CategoryFilesystem,
		Title:       models.LocalizedText{"en": "Find file"},
		Description: models.LocalizedText{"en": "Find files"},
	}, {
		ID:          "disk-usage",
		Category:    models.CategorySystem,
		Title:       models.LocalizedText{"en": "Disk usage"},
		Description: models.LocalizedText{"en": "Show disk usage"},
	}}, nil
}

// TestSearchModelSelection enters the selected category from All mode.
func TestSearchModelSelection(t *testing.T) {
	model, err := screens.NewSearchModel(fakeSearcher{}, "en", testStyles(), nil, nil, "linux-helper", "Search", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "type to search, left/right category, up/down move, enter open, ctrl+c quit")
	require.NoError(t, err)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	searchModel := updated.(screens.SearchModel)
	category, ok := (&searchModel).ConsumeCategorySelection()
	require.True(t, ok)
	assert.Equal(t, models.CategoryFilesystem, category)
}

// TestDetailModelBack pops on escape.
func TestDetailModelBack(t *testing.T) {
	model := screens.NewDetailModel(models.Recipe{Title: models.LocalizedText{"en": "Find file"}}, "en", testStyles(), false, "Run", "Back", "Favorite", "Remove", "Add")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	detailModel := updated.(screens.DetailModel)
	assert.True(t, (&detailModel).ConsumeBack())
}

// TestDetailModelExecute opens the form flow on enter.
func TestDetailModelExecute(t *testing.T) {
	model := screens.NewDetailModel(models.Recipe{Title: models.LocalizedText{"en": "Find file"}}, "en", testStyles(), false, "Run", "Back", "Favorite", "Remove", "Add")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	detailModel := updated.(screens.DetailModel)
	assert.True(t, (&detailModel).ConsumeExecute())
}

// TestFormModelSubmit collects field values.
func TestFormModelSubmit(t *testing.T) {
	model := screens.NewFormModel(models.Recipe{
		Title:       models.LocalizedText{"en": "Find file"},
		Description: models.LocalizedText{"en": "Find files"},
		Execution:   models.ExecutionTypeDirect,
		Binary:      "find",
		Args:        []string{"{{path}}"},
		Fields: []models.Field{{
			Name:     "path",
			Type:     models.FieldTypeString,
			Required: true,
			Default:  ".",
		}},
	}, "en", testStyles(), "Preview", "Submit", "Back")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	formModel := updated.(screens.FormModel)
	values, ok := (&formModel).ConsumeSubmit()
	require.True(t, ok)
	assert.Equal(t, ".", values["path"])
	assert.Equal(t, "find .", formModel.Preview())
}

// TestSearchModelFavoritesPrioritizesFavorites renders favorites first.
func TestSearchModelFavoritesPrioritizesFavorites(t *testing.T) {
	model, err := screens.NewSearchModel(fakeSearcher{}, "en", testStyles(), []string{"find-file"}, nil, "linux-helper", "Search", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "type to search, left/right category, up/down move, enter open, ctrl+c quit")
	require.NoError(t, err)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	searchModel := updated.(screens.SearchModel)
	searchModel.SetSelectedCategory(models.CategoryFilesystem)

	assert.Contains(t, searchModel.View(), "[*] Find file")
}

// TestSearchModelRecentCommands renders recent command history.
func TestSearchModelRecentCommands(t *testing.T) {
	model, err := screens.NewSearchModel(fakeSearcher{}, "en", testStyles(), nil, []string{"find .", "du -sh /var"}, "linux-helper", "Search", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "type to search, left/right category, up/down move, enter open, ctrl+c quit")
	require.NoError(t, err)

	view := model.View()
	assert.Contains(t, view, "Recent commands")
	assert.Contains(t, view, "- find .")
	assert.Contains(t, view, "- du -sh /var")
}

// TestSearchModelKeyboardShortcutsSupportProductiveNavigation.
func TestSearchModelKeyboardShortcutsSupportProductiveNavigation(t *testing.T) {
	model, err := screens.NewSearchModel(fakeSearcher{}, "en", testStyles(), nil, nil, "linux-helper", "Search", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "type to search, left/right category, up/down move, enter open, ctrl+c quit")
	require.NoError(t, err)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Contains(t, updated.View(), "Category:")
	assert.Contains(t, updated.View(), "[System]")
}

// TestSearchModelGroupsAndFiltersByCategory renders category sections and filters.
func TestSearchModelGroupsAndFiltersByCategory(t *testing.T) {
	model, err := screens.NewSearchModel(fakeSearcher{}, "en", testStyles(), nil, nil, "linux-helper", "Search", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "type to search, left/right category, up/down move, enter open, ctrl+c quit")
	require.NoError(t, err)

	view := model.View()
	assert.Contains(t, view, "Filesystem (1)")
	assert.Contains(t, view, "System (1)")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	filteredView := updated.View()
	assert.Contains(t, filteredView, "Find file")
	assert.NotContains(t, filteredView, "Disk usage")
}

// TestSearchModelTypingUpdatesQuery ensures search input remains usable.
func TestSearchModelTypingUpdatesQuery(t *testing.T) {
	model, err := screens.NewSearchModel(fakeSearcher{}, "en", testStyles(), nil, nil, "linux-helper", "Search", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "type to search, left/right category, up/down move, enter open, ctrl+c quit")
	require.NoError(t, err)

	updated, _ := model.Update(tea.KeyMsg{Runes: []rune{'f'}, Type: tea.KeyRunes})
	assert.Contains(t, updated.View(), "> f")
}

// TestDetailModelRunShortcutUsesR.
func TestDetailModelRunShortcutUsesR(t *testing.T) {
	model := screens.NewDetailModel(models.Recipe{Title: models.LocalizedText{"en": "Find file"}}, "en", testStyles(), false, "Run", "Back", "Favorite", "Remove", "Add")
	updated, _ := model.Update(tea.KeyMsg{Runes: []rune{'r'}, Type: tea.KeyRunes})
	detailModel := updated.(screens.DetailModel)
	assert.True(t, (&detailModel).ConsumeExecute())
}

// TestFormModelSubmitShortcutUsesCtrlS.
func TestFormModelSubmitShortcutUsesCtrlS(t *testing.T) {
	model := screens.NewFormModel(models.Recipe{
		Title:       models.LocalizedText{"en": "Find file"},
		Description: models.LocalizedText{"en": "Find files"},
		Execution:   models.ExecutionTypeDirect,
		Binary:      "find",
		Args:        []string{"{{path}}"},
		Fields: []models.Field{{
			Name:     "path",
			Type:     models.FieldTypeString,
			Required: true,
			Default:  ".",
		}},
	}, "en", testStyles(), "Preview", "Submit", "Back")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	formModel := updated.(screens.FormModel)
	values, ok := (&formModel).ConsumeSubmit()
	require.True(t, ok)
	assert.Equal(t, ".", values["path"])
}

// TestDetailModelToggleFavorite marks the requested change.
func TestDetailModelToggleFavorite(t *testing.T) {
	model := screens.NewDetailModel(models.Recipe{Title: models.LocalizedText{"en": "Find file"}}, "en", testStyles(), false, "Run", "Back", "Favorite", "Remove", "Add")
	updated, _ := model.Update(tea.KeyMsg{Runes: []rune{'f'}, Type: tea.KeyRunes})
	detailModel := updated.(screens.DetailModel)

	assert.True(t, (&detailModel).ConsumeToggleFavorite())
}
