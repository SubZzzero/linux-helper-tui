package screens_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

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
	}, {
		ID:          "port-owner",
		Category:    models.CategoryTroubleshooting,
		Title:       models.LocalizedText{"en": "Find port owner"},
		Description: models.LocalizedText{"en": "Find the process behind a port"},
	}}
}

func testStyles() uitheme.Styles {
	return uitheme.NewStyles(uitheme.Definition{Name: "test", BorderColor: "63", AccentColor: "213"})
}

// TestCatalogModelSelection enters the selected category from the category list.
func TestCatalogModelSelection(t *testing.T) {
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")

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
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), []string{"find-file"}, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	catalogModel := updated.(screens.CatalogModel)
	catalogModel.SetSelectedCategory(models.CategoryFilesystem)

	assert.Contains(t, catalogModel.View(), "[*] Find file")
}

// TestCatalogModelRecentCommands renders recent command history.
func TestCatalogModelRecentCommands(t *testing.T) {
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), nil, []string{"find .", "du -sh /var"}, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")

	view := model.View()
	assert.Contains(t, view, "Recent commands")
	assert.Contains(t, view, "- find .")
	assert.Contains(t, view, "- du -sh /var")
}

// TestCatalogModelBackReturnsToCategories uses escape to leave a category.
func TestCatalogModelBackReturnsToCategories(t *testing.T) {
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	catalogModel := updated.(screens.CatalogModel)
	category, ok := (&catalogModel).ConsumeCategorySelection()
	require.True(t, ok)
	assert.Equal(t, models.CategorySystem, category)

	catalogModel.SetSelectedCategory(category)

	updated, _ = catalogModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	view := updated.View()

	assert.Contains(t, findLineContaining(view, "Filesystem"), "Files, directories, and permissions")
	assert.Contains(t, findLineContaining(view, "System"), "System, disks, and resources")
	assert.Contains(t, findLineContaining(view, "Troubleshooting"), "Failure triage, diagnostics, and root-cause checks")
	assert.Contains(t, findLineContaining(view, "System"), ">")
	assert.NotContains(t, findLineContaining(view, "Filesystem"), ">")
	assert.NotContains(t, view, "[*] Find file")
}

// TestCatalogModelGroupsAndFiltersByCategory renders category list and selected recipes.
func TestCatalogModelGroupsAndFiltersByCategory(t *testing.T) {
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")

	view := model.View()
	assert.Contains(t, findLineContaining(view, "Filesystem"), "Files, directories, and permissions")
	assert.Contains(t, findLineContaining(view, "System"), "System, disks, and resources")
	assert.Contains(t, findLineContaining(view, "Troubleshooting"), "Failure triage, diagnostics, and root-cause checks")
	assert.NotContains(t, view, "Category:")
	assert.NotContains(t, view, "(1)")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	catalogModel := updated.(screens.CatalogModel)
	category, ok := (&catalogModel).ConsumeCategorySelection()
	require.True(t, ok)
	catalogModel.SetSelectedCategory(category)
	filteredView := catalogModel.View()
	assert.Contains(t, filteredView, "Find file")
	assert.NotContains(t, filteredView, "Disk usage")
	assert.NotContains(t, filteredView, "[filesystem]")
}

// TestCatalogModelCategorySelectionResetsRecipeIndex opens the first recipe in a category.
func TestCatalogModelCategorySelectionResetsRecipeIndex(t *testing.T) {
	model := screens.NewCatalogModel([]models.Recipe{{
		ID:          "find-file",
		Category:    models.CategoryFilesystem,
		Title:       models.LocalizedText{"en": "Find file"},
		Description: models.LocalizedText{"en": "Find files"},
	}, {
		ID:          "disk-usage",
		Category:    models.CategorySystem,
		Title:       models.LocalizedText{"en": "Disk usage"},
		Description: models.LocalizedText{"en": "Show disk usage"},
	}, {
		ID:          "show-memory",
		Category:    models.CategorySystem,
		Title:       models.LocalizedText{"en": "Show memory"},
		Description: models.LocalizedText{"en": "Show memory usage"},
	}}, "en", testStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	catalogModel := updated.(screens.CatalogModel)
	category, ok := (&catalogModel).ConsumeCategorySelection()
	require.True(t, ok)
	assert.Equal(t, models.CategorySystem, category)

	catalogModel.SetSelectedCategory(category)
	updated, _ = catalogModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	catalogModel = updated.(screens.CatalogModel)
	recipe, ok := (&catalogModel).ConsumeSelection()
	require.True(t, ok)
	assert.Equal(t, "disk-usage", recipe.ID)
}

// TestCatalogModelCategoryDescriptionsAlign keeps category descriptions in one column.
func TestCatalogModelCategoryDescriptionsAlign(t *testing.T) {
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")
	view := model.View()

	filesystemLine := findLineContaining(view, "Filesystem")
	systemLine := findLineContaining(view, "System")
	troubleshootingLine := findLineContaining(view, "Troubleshooting")
	require.NotEmpty(t, filesystemLine)
	require.NotEmpty(t, systemLine)
	require.NotEmpty(t, troubleshootingLine)

	assert.Equal(t, strings.Index(filesystemLine, "Files, directories, and permissions"), strings.Index(systemLine, "System, disks, and resources"))
	assert.Equal(t, strings.Index(filesystemLine, "Files, directories, and permissions"), strings.Index(troubleshootingLine, "Failure triage, diagnostics, and root-cause checks"))
}

// TestCatalogModelFitsNarrowTerminal avoids forcing a wide frame in small terminals.
func TestCatalogModelFitsNarrowTerminal(t *testing.T) {
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 30, Height: 12})
	view := updated.View()

	assert.LessOrEqual(t, maxLineWidth(view), 30)
	assert.LessOrEqual(t, lineCount(view), 12)
	assert.Contains(t, view, "┌")
	assert.Contains(t, view, "┘")
}

// TestCatalogModelWrapsLongRecentCommands keeps long commands inside the frame.
func TestCatalogModelWrapsLongRecentCommands(t *testing.T) {
	model := screens.NewCatalogModel(
		testRecipes(),
		"en",
		testStyles(),
		nil,
		[]string{"systemctl list-dependencies multi-user.target --no-pager"},
		"linux-helper",
		"Empty",
		"Recent commands",
		"No recent commands yet.",
		"up/down move, enter open, esc back, ctrl+c quit",
	)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 34, Height: 16})
	view := updated.View()

	assert.Contains(t, view, "- systemctl")
	assert.Contains(t, view, "  multi-user.target --no-pager")
	assert.LessOrEqual(t, maxLineWidth(view), 34)
}

// TestCatalogModelTypingDoesNotChangeView keeps browse-only input inactive.
func TestCatalogModelTypingDoesNotChangeView(t *testing.T) {
	model := screens.NewCatalogModel(testRecipes(), "en", testStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")

	before := model.View()
	updated, _ := model.Update(tea.KeyMsg{Runes: []rune{'f'}, Type: tea.KeyRunes})
	assert.Equal(t, before, updated.View())
}

// TestSharedFooterRendersGitHubCredit keeps the application credit visible.
func TestSharedFooterRendersGitHubCredit(t *testing.T) {
	model := screens.NewDetailModel(models.Recipe{Title: models.LocalizedText{"en": "Find file"}}, "en", testStyles(), false, "Run", "Back", "Favorite", "Remove", "Add")

	assert.Contains(t, model.View(), "Developed by github.com/SubZzzero")
}

// TestDetailModelRuneRDoesNotTriggerExecute keeps text runes from acting as shortcuts.
func TestDetailModelRuneRDoesNotTriggerExecute(t *testing.T) {
	model := screens.NewDetailModel(models.Recipe{Title: models.LocalizedText{"en": "Find file"}}, "en", testStyles(), false, "Run", "Back", "Favorite", "Remove", "Add")
	updated, _ := model.Update(tea.KeyMsg{Runes: []rune{'r'}, Type: tea.KeyRunes})
	detailModel := updated.(screens.DetailModel)
	assert.False(t, (&detailModel).ConsumeExecute())
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
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
	detailModel := updated.(screens.DetailModel)

	assert.True(t, (&detailModel).ConsumeToggleFavorite())
}

// TestDetailModelRuneFDoesNotToggleFavorite keeps text runes available for input.
func TestDetailModelRuneFDoesNotToggleFavorite(t *testing.T) {
	model := screens.NewDetailModel(models.Recipe{Title: models.LocalizedText{"en": "Find file"}}, "en", testStyles(), false, "Run", "Back", "Favorite", "Remove", "Add")
	updated, _ := model.Update(tea.KeyMsg{Runes: []rune{'f'}, Type: tea.KeyRunes})
	detailModel := updated.(screens.DetailModel)

	assert.False(t, (&detailModel).ConsumeToggleFavorite())
}

// TestResultModelScrollsLargeOutput keeps large command output accessible in a viewport.
func TestResultModelScrollsLargeOutput(t *testing.T) {
	model := screens.NewResultModel(
		models.Recipe{Title: models.LocalizedText{"en": "Process list"}},
		"en",
		testStyles(),
		"Running",
		"Stop",
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

// TestResultModelRunningEscapeRequestsStop uses escape to interrupt a running command.
func TestResultModelRunningEscapeRequestsStop(t *testing.T) {
	model := screens.NewResultModel(
		models.Recipe{Title: models.LocalizedText{"en": "Follow logs"}},
		"en",
		testStyles(),
		"Running",
		"Stop",
		"Done",
		"Back",
		"Scroll",
	)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	resultModel := updated.(screens.ResultModel)
	assert.True(t, (&resultModel).ConsumeStop())
	assert.False(t, (&resultModel).ConsumeBack())
}

// TestResultModelShowsLiveOutputWhileRunning keeps streamed text visible in a sized viewport.
func TestResultModelShowsLiveOutputWhileRunning(t *testing.T) {
	model := screens.NewResultModel(
		models.Recipe{Title: models.LocalizedText{"en": "Follow logs"}},
		"en",
		testStyles(),
		"Running",
		"Stop",
		"Done",
		"Back",
		"Scroll",
	)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 16})
	resultModel := updated.(screens.ResultModel)
	resultModel.AppendOutput("stdout", "line one\n")
	resultModel.AppendOutput("stderr", "warn\n")
	view := resultModel.View()

	assert.Contains(t, view, "line one")
	assert.Contains(t, view, "warn")
	assert.Contains(t, view, "Running")
	assert.NotContains(t, view, "Execution finished")
}

func findLineContaining(view string, substring string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, substring) {
			return line
		}
	}

	return ""
}

func maxLineWidth(view string) int {
	width := 0
	for _, line := range strings.Split(view, "\n") {
		width = maxInt(width, utf8.RuneCountInString(line))
	}

	return width
}

func lineCount(view string) int {
	return len(strings.Split(view, "\n"))
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}

	return right
}
