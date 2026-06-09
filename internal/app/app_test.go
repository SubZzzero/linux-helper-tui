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

type fakeSearcher struct{}

type fakeExecutor struct {
	result models.ExecutionResult
	err    error
	called int
}

// Execute records one execution request.
func (e *fakeExecutor) Execute(_ context.Context, _ models.Recipe, _ map[string]string, _ bool) (models.ExecutionResult, error) {
	e.called++
	return e.result, e.err
}

func appTestStyles() uitheme.Styles {
	return uitheme.NewStyles(uitheme.Definition{Name: "test", BorderColor: "63", AccentColor: "213"})
}

// Search returns one static recipe.
func (fakeSearcher) Search(query string) ([]models.Recipe, error) {
	return []models.Recipe{{
		ID:          "find-file",
		Category:    models.CategoryFilesystem,
		Risk:        models.RiskSafe,
		Execution:   models.ExecutionTypeDirect,
		Binary:      "find",
		Args:        []string{"{{path}}"},
		Fields:      []models.Field{{Name: "path", Type: models.FieldTypeString, Required: true, Default: "."}},
		Title:       models.LocalizedText{"en": "Find file"},
		Description: models.LocalizedText{"en": "Find files"},
	}}, nil
}

// TestModelView renders the active screen.
func TestModelView(t *testing.T) {
	searchModel, err := screens.NewSearchModel(fakeSearcher{}, "en", appTestStyles(), "linux-helper", "Search", "Empty")
	require.NoError(t, err)

	model := app.NewModel(searchModel, "en", appTestStyles(), nil, nil)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotEmpty(t, updated.View())
}

// TestModelExecutesSafeRecipe drives the minimal execution flow.
func TestModelExecutesSafeRecipe(t *testing.T) {
	searchModel, err := screens.NewSearchModel(fakeSearcher{}, "en", appTestStyles(), "linux-helper", "Search", "Empty")
	require.NoError(t, err)

	executor := &fakeExecutor{result: models.ExecutionResult{Command: "find .", ExitCode: 0, Stdout: "ok"}}
	model := app.NewModel(searchModel, "en", appTestStyles(), executor, nil)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
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
}
