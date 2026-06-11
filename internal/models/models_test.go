package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"linux-helper/internal/models"
)

// TestRecipeValidate validates a well-formed recipe.
func TestRecipeValidate(t *testing.T) {
	recipe := models.Recipe{
		ID:        "find-file",
		Version:   1,
		Type:      "recipe",
		Category:  models.CategoryFilesystem,
		Risk:      models.RiskSafe,
		Execution: models.ExecutionTypeDirect,
		Binary:    "find",
		Title:     models.LocalizedText{"en": "Find file"},
		Fields:    []models.Field{{Name: "path", Type: models.FieldTypeString}},
	}

	require.NoError(t, recipe.Validate())
	assert.Equal(t, "Find file", recipe.Title.Resolve("en"))
}

// TestParseCategory accepts known categories and rejects unknown ones.
func TestParseCategory(t *testing.T) {
	category, err := models.ParseCategory("filesystem")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryFilesystem, category)

	category, err = models.ParseCategory("network")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryNetwork, category)

	category, err = models.ParseCategory("environment")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryEnvironment, category)

	category, err = models.ParseCategory("text")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryText, category)

	category, err = models.ParseCategory("users")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryUsers, category)

	_, err = models.ParseCategory("missing")
	assert.Error(t, err)
}

// TestCategoryDisplayName returns stable labels for known categories.
func TestCategoryDisplayName(t *testing.T) {
	assert.Equal(t, "Filesystem", models.CategoryFilesystem.DisplayName())
	assert.Equal(t, "Environment", models.CategoryEnvironment.DisplayName())
	assert.Equal(t, "Network", models.CategoryNetwork.DisplayName())
	assert.Equal(t, "System", models.CategorySystem.DisplayName())
	assert.Equal(t, "Text", models.CategoryText.DisplayName())
	assert.Equal(t, "Users", models.CategoryUsers.DisplayName())
}
