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

// TestParseCategory rejects unknown categories.
func TestParseCategory(t *testing.T) {
	category, err := models.ParseCategory("filesystem")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryFilesystem, category)

	_, err = models.ParseCategory("missing")
	assert.Error(t, err)
}
