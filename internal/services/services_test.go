package services_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"linux-helper/internal/executor"
	"linux-helper/internal/models"
	"linux-helper/internal/services"
	"linux-helper/internal/storage"
)

type fakeLoader struct{}

// Load returns one static recipe.
func (fakeLoader) Load() ([]models.Recipe, error) {
	return []models.Recipe{{
		ID:        "find-file",
		Version:   1,
		Type:      "recipe",
		Category:  models.CategoryFilesystem,
		Risk:      models.RiskSafe,
		Execution: models.ExecutionTypeDirect,
		Binary:    "find",
		Title:     models.LocalizedText{"en": "Find file"},
	}}, nil
}

// TestFavoritesServiceToggle persists favorite state.
func TestFavoritesServiceToggle(t *testing.T) {
	store := storage.NewFavoritesStore(filepath.Join(t.TempDir(), "favorites.yaml"))
	service := services.NewFavoritesService(store)

	isFavorite, err := service.Toggle("find-file")
	require.NoError(t, err)
	assert.True(t, isFavorite)

	loaded, err := service.Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"find-file"}, loaded)

	isFavorite, err = service.Toggle("find-file")
	require.NoError(t, err)
	assert.False(t, isFavorite)

	loaded, err = service.Load()
	require.NoError(t, err)
	assert.Empty(t, loaded)
}

// TestRecentServiceLoad returns stored recent commands.
func TestRecentServiceLoad(t *testing.T) {
	store := storage.NewRecentStore(filepath.Join(t.TempDir(), "recent.yaml"))
	require.NoError(t, store.Save([]string{"find .", "du -sh /var"}))

	service := services.NewRecentService(store)
	loaded, err := service.Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"find .", "du -sh /var"}, loaded)
}

type fakeRunner struct{}

// Run returns a static execution result.
func (fakeRunner) Run(_ context.Context, _ string, _ ...string) (models.ExecutionResult, error) {
	return models.ExecutionResult{Command: "find ."}, nil
}

// RunShell returns a static execution result.
func (fakeRunner) RunShell(_ context.Context, _ string) (models.ExecutionResult, error) {
	return models.ExecutionResult{Command: "echo ok"}, nil
}

// RunStreaming returns a static execution result for streaming callers.
func (fakeRunner) RunStreaming(_ context.Context, _ string, sink executor.OutputSink, _ ...string) (models.ExecutionResult, error) {
	if sink != nil {
		sink("stdout", "out\n")
	}

	return models.ExecutionResult{Command: "tail -F /var/log/syslog", Stdout: "out\n"}, nil
}

// RunShellStreaming returns a static execution result for streaming callers.
func (fakeRunner) RunShellStreaming(_ context.Context, _ string, sink executor.OutputSink) (models.ExecutionResult, error) {
	if sink != nil {
		sink("stderr", "warn\n")
	}

	return models.ExecutionResult{Command: "echo ok", Stderr: "warn\n"}, nil
}

// TestRecipeServiceAll wires loader and registry correctly.
func TestRecipeServiceAll(t *testing.T) {
	recipeService, err := services.NewRecipeService(fakeLoader{})
	require.NoError(t, err)

	results := recipeService.All()
	assert.Len(t, results, 1)
	assert.Equal(t, "find-file", results[0].ID)
}

// TestExecutionService executes and records commands.
func TestExecutionService(t *testing.T) {
	recent := storage.NewRecentStore(filepath.Join(t.TempDir(), "recent.yaml"))
	service := services.NewExecutionService(fakeRunner{}, recent)

	result, err := service.Execute(context.Background(), models.Recipe{
		ID:        "find-file",
		Risk:      models.RiskSafe,
		Execution: models.ExecutionTypeDirect,
		Binary:    "find",
		Args:      []string{"{{path}}"},
	}, map[string]string{"path": "."}, true)

	require.NoError(t, err)
	assert.Equal(t, "find .", result.Command)
}

// TestExecutionServiceRequiresConfirmation blocks dangerous recipes.
func TestExecutionServiceRequiresConfirmation(t *testing.T) {
	service := services.NewExecutionService(fakeRunner{}, nil)

	_, err := service.Execute(context.Background(), models.Recipe{
		ID:        "rm-all",
		Risk:      models.RiskDangerous,
		Execution: models.ExecutionTypeDirect,
		Binary:    "rm",
		Args:      []string{"-rf", "{{path}}"},
	}, map[string]string{"path": "/tmp/demo"}, false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "confirmation required")
}

// TestExecutionServiceStreaming forwards live output and records the final command.
func TestExecutionServiceStreaming(t *testing.T) {
	recent := storage.NewRecentStore(filepath.Join(t.TempDir(), "recent.yaml"))
	service := services.NewExecutionService(fakeRunner{}, recent)
	chunks := []string{}

	result, err := service.ExecuteStreaming(context.Background(), models.Recipe{
		ID:        "follow-log-file",
		Risk:      models.RiskSafe,
		Execution: models.ExecutionTypeDirect,
		Binary:    "tail",
		Args:      []string{"-F", "{{path}}"},
	}, map[string]string{"path": "/var/log/syslog"}, true, func(stream string, chunk string) {
		chunks = append(chunks, stream+":"+chunk)
	})

	require.NoError(t, err)
	assert.Equal(t, "tail -F /var/log/syslog", result.Command)
	assert.Equal(t, []string{"stdout:out\n"}, chunks)
	loaded, loadErr := recent.Load()
	require.NoError(t, loadErr)
	assert.Equal(t, []string{"tail -F /var/log/syslog"}, loaded)
}
