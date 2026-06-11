package screens_test

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"linux-helper/internal/models"
	"linux-helper/internal/tui/screens"
	uitheme "linux-helper/internal/tui/theme"
)

func testRecipes() []models.Recipe {
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
	}}
}

func testStyles() uitheme.Styles {
	return uitheme.NewStyles(uitheme.Definition{Name: "test", BorderColor: "63", AccentColor: "213"})
}

// TestCatalogModelSelection enters the selected category from All mode.
func TestCatalogModelSelection(t *testing.T) {
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "left/right category, up/down move, enter open, ctrl+c quit")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	catalogModel := updated.(screens.CatalogModel)
	category, ok := (&catalogModel).ConsumeCategorySelection()
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

// TestCatalogModelFavoritesPrioritizesFavorites renders favorites first.
func TestCatalogModelFavoritesPrioritizesFavorites(t *testing.T) {
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), []string{"find-file"}, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "left/right category, up/down move, enter open, ctrl+c quit")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	catalogModel := updated.(screens.CatalogModel)
	catalogModel.SetSelectedCategory(models.CategoryFilesystem)

	assert.Contains(t, catalogModel.View(), "[*] Find file")
}

// TestCatalogModelRecentCommands renders recent command history.
func TestCatalogModelRecentCommands(t *testing.T) {
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), nil, []string{"find .", "du -sh /var"}, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "left/right category, up/down move, enter open, ctrl+c quit")

	view := model.View()
	assert.Contains(t, view, "Recent commands")
	assert.Contains(t, view, "- find .")
	assert.Contains(t, view, "- du -sh /var")
}

// TestCatalogModelKeyboardShortcutsSupportProductiveNavigation.
func TestCatalogModelKeyboardShortcutsSupportProductiveNavigation(t *testing.T) {
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "left/right category, up/down move, enter open, ctrl+c quit")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Contains(t, updated.View(), "Category:")
	assert.Contains(t, updated.View(), "[System]")
}

// TestCatalogModelGroupsAndFiltersByCategory renders category sections and filters.
func TestCatalogModelGroupsAndFiltersByCategory(t *testing.T) {
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "left/right category, up/down move, enter open, ctrl+c quit")

	view := model.View()
	assert.Contains(t, view, "Filesystem (1)")
	assert.Contains(t, view, "System (1)")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	filteredView := updated.View()
	assert.Contains(t, filteredView, "Find file")
	assert.NotContains(t, filteredView, "Disk usage")
}

// TestCatalogModelTypingDoesNotChangeView keeps browse-only input inactive.
func TestCatalogModelTypingDoesNotChangeView(t *testing.T) {
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "left/right category, up/down move, enter open, ctrl+c quit")

	before := model.View()
	updated, _ := model.Update(tea.KeyMsg{Runes: []rune{'f'}, Type: tea.KeyRunes})
	assert.Equal(t, before, updated.View())
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

// TestResultModelScrollsLargeOutput keeps large command output accessible in a viewport.
func TestResultModelScrollsLargeOutput(t *testing.T) {
	model := screens.NewResultModel(
		models.Recipe{Title: models.LocalizedText{"en": "Process list"}},
		"en",
		testStyles(),
		"Running",
		"Done",
		"Back",
		"Scroll",
	)

	outputLines := make([]string, 0, 20)
	for index := 1; index <= 20; index++ {
		outputLines = append(outputLines, fmt.Sprintf("line %02d", index))
	}

	model.SetOutcome(models.ExecutionResult{
		Command:  "ps aux",
		ExitCode: 0,
		Stdout:   strings.Join(outputLines, "\n"),
	}, nil)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 16})
	resultModel := updated.(screens.ResultModel)

	initialView := resultModel.View()
	assert.Contains(t, initialView, "line 01")
	assert.NotContains(t, initialView, "line 20")

	for step := 0; step < 5; step++ {
		updated, _ = resultModel.Update(tea.KeyMsg{Type: tea.KeyPgDown})
		resultModel = updated.(screens.ResultModel)
	}

	assert.NotContains(t, resultModel.View(), "line 01")
	assert.Contains(t, resultModel.View(), "line 20")
	assert.False(t, (&resultModel).ConsumeBack())
	updated, _ = resultModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	resultModel = updated.(screens.ResultModel)
	assert.True(t, (&resultModel).ConsumeBack())
}
