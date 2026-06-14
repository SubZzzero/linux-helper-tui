package app_test

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"linux-helper/internal/app"
	"linux-helper/internal/models"
	"linux-helper/internal/storage"
)

// TestModelLocaleHotkeyUpdatesDangerousConfirmWithoutLosingPreview keeps confirmation state stable.
func TestModelLocaleHotkeyUpdatesDangerousConfirmWithoutLosingPreview(t *testing.T) {
	executor := &fakeExecutor{result: models.ExecutionResult{Command: "rm -rf /tmp/cache", ExitCode: 0, Stdout: "done"}}
	catalogModel := appTestCatalog(appTestRecipes(), nil)
	model := app.NewModel(catalogModel, "en", appTestStyles(), nil, nil, nil, executor, nil)
	saved := []storage.Config{}
	configureAppPreferences(&model, &saved)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})

	updatedModel := updated.(app.Model)
	updated, cmd := updatedModel.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	require.NotNil(t, cmd)
	updated, _ = updated.Update(cmd())
	view := updated.View()

	assert.Contains(t, view, "Підтвердити UA")
	assert.Contains(t, view, "Видалити дерево")
	assert.Contains(t, view, "rm -rf /tmp/cache")
	require.Len(t, saved, 1)
	assert.Equal(t, storage.Config{Locale: "ua", Theme: "dark"}, saved[0])
	assert.Equal(t, 0, executor.callCount())

	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	updated, _ = updated.Update(cmd())
	assert.Equal(t, 1, executor.callCount())
	assert.Contains(t, updated.View(), "Execution finished")
	assert.Contains(t, updated.View(), "rm -rf /tmp/cache")
}

// TestModelDangerousConfirmBackClearsPendingExecution prevents stale execution after cancel.
func TestModelDangerousConfirmBackClearsPendingExecution(t *testing.T) {
	executor := &fakeExecutor{result: models.ExecutionResult{Command: "rm -rf /tmp/cache", ExitCode: 0, Stdout: "done"}}
	catalogModel := appTestCatalog(appTestRecipes(), nil)
	model := app.NewModel(catalogModel, "en", appTestStyles(), nil, nil, nil, executor, nil)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Contains(t, updated.View(), "Confirmation required")

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	view := updated.View()
	assert.Contains(t, view, "Preview")
	assert.Contains(t, view, "rm -rf /tmp/cache")
	assert.Equal(t, 0, executor.callCount())

	_, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Nil(t, cmd)
	assert.Equal(t, 0, executor.callCount())
}

// TestModelThemeHotkeyPersistsConfig keeps locale stable while cycling the theme.
func TestModelThemeHotkeyPersistsConfig(t *testing.T) {
	catalogModel := appTestCatalog(appTestRecipes(), nil)
	model := app.NewModel(catalogModel, "en", appTestStyles(), nil, nil, nil, nil, nil)
	saved := []storage.Config{}
	configureAppPreferences(&model, &saved)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel := updated.(app.Model)
	updated, cmd := updatedModel.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	require.NotNil(t, cmd)
	updated, _ = updated.Update(cmd())
	view := updated.View()

	require.Len(t, saved, 1)
	assert.Equal(t, storage.Config{Locale: "en", Theme: "light"}, saved[0])
	assert.Contains(t, view, "Find file")
	assert.Contains(t, view, "Detail EN")
	assert.Contains(t, view, "Favorite")
}

// TestModelThemeHotkeyOnResultScreenPreservesOutput keeps result text visible after a refresh.
func TestModelThemeHotkeyOnResultScreenPreservesOutput(t *testing.T) {
	outputLines := make([]string, 0, 20)
	for index := 1; index <= 20; index++ {
		outputLines = append(outputLines, fmt.Sprintf("line %02d", index))
	}

	executor := &fakeExecutor{result: models.ExecutionResult{Command: "ps aux", ExitCode: 0, Stdout: strings.Join(outputLines, "\n")}}
	catalogModel := appTestCatalog(appTestRecipes(), nil)
	model := app.NewModel(catalogModel, "en", appTestStyles(), nil, nil, nil, executor, nil)
	saved := []storage.Config{}
	configureAppPreferences(&model, &saved)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 16})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	updated, _ = updated.Update(cmd())

	for step := 0; step < 5; step++ {
		updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	}

	require.Contains(t, updated.View(), "line 20")
	updatedModel := updated.(app.Model)
	updated, cmd = updatedModel.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	require.NotNil(t, cmd)
	updated, _ = updated.Update(cmd())
	assert.Contains(t, updated.View(), "line 20")
	assert.Contains(t, updated.View(), "Scroll EN")
	require.Len(t, saved, 1)
	assert.Equal(t, "light", saved[0].Theme)
}
