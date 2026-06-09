package services_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

type fakeRunner struct{}

// Run returns a static execution result.
func (fakeRunner) Run(_ context.Context, _ string, _ ...string) (models.ExecutionResult, error) {
	return models.ExecutionResult{Command: "find ."}, nil
}

// RunShell returns a static execution result.
func (fakeRunner) RunShell(_ context.Context, _ string) (models.ExecutionResult, error) {
	return models.ExecutionResult{Command: "echo ok"}, nil
}

// TestRecipeAndSearchService wires loader and search correctly.
func TestRecipeAndSearchService(t *testing.T) {
	recipeService, err := services.NewRecipeService(fakeLoader{})
	require.NoError(t, err)

	searchService := services.NewSearchService(recipeService.All())
	results, err := searchService.Search("find")
	require.NoError(t, err)
	assert.Len(t, results, 1)
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
