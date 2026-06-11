package app_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"linux-helper/internal/app"
	"linux-helper/internal/models"
	"linux-helper/internal/tui/screens"
	uitheme "linux-helper/internal/tui/theme"
)

func appTestRecipes() []models.Recipe {
	return []models.Recipe{{
		ID:          "find-file",
		Version:     1,
		Category:    models.CategoryFilesystem,
		Risk:        models.RiskSafe,
		Execution:   models.ExecutionTypeDirect,
		Binary:      "find",
		Args:        []string{"{{path}}"},
		Fields:      []models.Field{{Name: "path", Type: models.FieldTypeString, Required: true, Default: "."}},
		Title:       models.LocalizedText{"en": "Find file"},
		Description: models.LocalizedText{"en": "Find files"},
	}}
}

type fakeExecutor struct {
	result models.ExecutionResult
	err    error
	called int
}

type fakeFavorites struct {
	ids map[string]struct{}
}

type fakeRecent struct {
	commands []string
}

// Execute records one execution request.
func (e *fakeExecutor) Execute(_ context.Context, _ models.Recipe, _ map[string]string, _ bool) (models.ExecutionResult, error) {
	e.called++
	return e.result, e.err
}

// Load returns the current favorite identifiers.
func (f *fakeFavorites) Load() ([]string, error) {
	ids := make([]string, 0, len(f.ids))
	for recipeID := range f.ids {
		ids = append(ids, recipeID)
	}

	return ids, nil
}

// Toggle flips one recipe identifier.
func (f *fakeFavorites) Toggle(recipeID string) (bool, error) {
	if _, ok := f.ids[recipeID]; ok {
		delete(f.ids, recipeID)
		return false, nil
	}

	f.ids[recipeID] = struct{}{}
	return true, nil
}

// Load returns the recent commands.
func (f *fakeRecent) Load() ([]string, error) {
	return append([]string(nil), f.commands...), nil
}

func appTestStyles() uitheme.Styles {
	return uitheme.NewStyles(uitheme.Definition{Name: "test", BorderColor: "63", AccentColor: "213"})
}

// TestModelView renders the active screen.
func TestModelView(t *testing.T) {
	catalogModel := screens.NewCatalogModel(appTestRecipes(), "en", appTestStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "left/right category, up/down move, enter open, ctrl+c quit")
	model := app.NewModel(catalogModel, "en", appTestStyles(), nil, nil, nil, nil, nil)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotEmpty(t, updated.View())
}

// TestModelExecutesSafeRecipe drives the minimal execution flow.
func TestModelExecutesSafeRecipe(t *testing.T) {
	executor := &fakeExecutor{result: models.ExecutionResult{Command: "find .", ExitCode: 0, Stdout: "ok"}}
	recent := &fakeRecent{commands: []string{"find ."}}
	catalogModel := screens.NewCatalogModel(appTestRecipes(), "en", appTestStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "left/right category, up/down move, enter open, ctrl+c quit")
	model := app.NewModel(catalogModel, "en", appTestStyles(), nil, recent, nil, executor, nil)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, updated)
	require.Nil(t, cmd)

	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, updated)
	require.Nil(t, cmd)

	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, updated)
	require.Nil(t, cmd)

	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, updated)
	require.NotNil(t, cmd)
	assert.Equal(t, 0, executor.called)

	updated, _ = updated.Update(cmd())
	assert.Equal(t, 1, executor.called)
	assert.Contains(t, updated.View(), "Execution finished")
	assert.Contains(t, updated.View(), "find .")

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Contains(t, updated.View(), "Recent commands")
	assert.Contains(t, updated.View(), "- find .")
}

// TestModelTogglesFavorites updates detail and catalog state.
func TestModelTogglesFavorites(t *testing.T) {
	favorites := &fakeFavorites{ids: map[string]struct{}{}}
	catalogModel := screens.NewCatalogModel(appTestRecipes(), "en", appTestStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "left/right category, up/down move, enter open, ctrl+c quit")
	model := app.NewModel(catalogModel, "en", appTestStyles(), favorites, nil, nil, nil, nil)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Runes: []rune{'f'}, Type: tea.KeyRunes})

	assert.Contains(t, updated.View(), "Favorite: yes")

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Contains(t, updated.View(), "[*] Find file")
}

// TestModelCarriesWindowSizeToResultScreen keeps result output usable without a manual resize.
func TestModelCarriesWindowSizeToResultScreen(t *testing.T) {
	executor := &fakeExecutor{result: models.ExecutionResult{Command: "ps aux", ExitCode: 0, Stdout: "header\nbody\nfooter"}}
	catalogModel := screens.NewCatalogModel(appTestRecipes(), "en", appTestStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "Category:", "All", "left/right category, up/down move, enter open, ctrl+c quit")
	model := app.NewModel(catalogModel, "en", appTestStyles(), nil, nil, nil, executor, nil)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	updated, _ = updated.Update(cmd())
	view := updated.View()

	assert.Contains(t, view, "Execution finished")
	assert.Contains(t, view, "header")
	assert.Contains(t, view, "body")
	assert.Contains(t, view, "footer")
	assert.Contains(t, view, "Use up/down or pgup/pgdn to scroll")
}
