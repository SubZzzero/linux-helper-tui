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

	category, err = models.ParseCategory("docker")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryDocker, category)

	category, err = models.ParseCategory("environment")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryEnvironment, category)

	category, err = models.ParseCategory("logs")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryLogs, category)

	category, err = models.ParseCategory("text")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryText, category)

	category, err = models.ParseCategory("packages")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryPackages, category)

	category, err = models.ParseCategory("processes")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryProcesses, category)

	category, err = models.ParseCategory("services")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryServices, category)

	category, err = models.ParseCategory("troubleshooting")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryTroubleshooting, category)

	category, err = models.ParseCategory("users")
	require.NoError(t, err)
	assert.Equal(t, models.CategoryUsers, category)

	_, err = models.ParseCategory("missing")
	assert.Error(t, err)
}

// TestCategoryDisplayName returns stable labels for known categories.
func TestCategoryDisplayName(t *testing.T) {
	assert.Equal(t, "Docker", models.CategoryDocker.DisplayName())
	assert.Equal(t, "Filesystem", models.CategoryFilesystem.DisplayName())
	assert.Equal(t, "Environment", models.CategoryEnvironment.DisplayName())
	assert.Equal(t, "Logs", models.CategoryLogs.DisplayName())
	assert.Equal(t, "Network", models.CategoryNetwork.DisplayName())
	assert.Equal(t, "Packages", models.CategoryPackages.DisplayName())
	assert.Equal(t, "Processes", models.CategoryProcesses.DisplayName())
	assert.Equal(t, "Services", models.CategoryServices.DisplayName())
	assert.Equal(t, "System", models.CategorySystem.DisplayName())
	assert.Equal(t, "Text", models.CategoryText.DisplayName())
	assert.Equal(t, "Troubleshooting", models.CategoryTroubleshooting.DisplayName())
	assert.Equal(t, "Users", models.CategoryUsers.DisplayName())
	assert.Equal(t, "custom", models.Category("custom").DisplayName())
}

// TestLocalizedTextResolve falls back across locale variants.
func TestLocalizedTextResolve(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Ukrainian", models.LocalizedText{"ua": "Ukrainian", "en": "English"}.Resolve("ua"))
	assert.Equal(t, "English", models.LocalizedText{"en": "English"}.Resolve("ru"))
	assert.Equal(t, "Fallback", models.LocalizedText{"ua": "", "ru": "Fallback"}.Resolve("ua"))
	assert.Equal(t, "", models.LocalizedText{}.Resolve("en"))
}

// TestFieldValid validates required and supported field metadata.
func TestFieldValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		field   models.Field
		wantErr string
	}{
		{
			name:    "missing name",
			field:   models.Field{Type: models.FieldTypeString},
			wantErr: "field name is required",
		},
		{
			name:    "missing type",
			field:   models.Field{Name: "path"},
			wantErr: "field \"path\" type is required",
		},
		{
			name:    "unsupported type",
			field:   models.Field{Name: "path", Type: models.FieldType("number")},
			wantErr: "field \"path\" type \"number\" is not supported",
		},
		{
			name:  "valid",
			field: models.Field{Name: "path", Type: models.FieldTypeString},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.field.Valid()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}

			require.EqualError(t, err, tt.wantErr)
		})
	}
}

// TestRecipeValidateErrors covers invalid recipe branches.
func TestRecipeValidateErrors(t *testing.T) {
	t.Parallel()

	validRecipe := models.Recipe{
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

	tests := []struct {
		name    string
		mutate  func(recipe *models.Recipe)
		wantErr string
	}{
		{name: "missing id", mutate: func(recipe *models.Recipe) { recipe.ID = "" }, wantErr: "recipe id is required"},
		{name: "invalid version", mutate: func(recipe *models.Recipe) { recipe.Version = 0 }, wantErr: "recipe \"find-file\" version must be positive"},
		{name: "invalid type", mutate: func(recipe *models.Recipe) { recipe.Type = "command" }, wantErr: "recipe \"find-file\" type must be recipe"},
		{name: "invalid category", mutate: func(recipe *models.Recipe) { recipe.Category = models.Category("invalid") }, wantErr: "recipe \"find-file\" category is invalid"},
		{name: "invalid risk", mutate: func(recipe *models.Recipe) { recipe.Risk = models.RiskLevel("invalid") }, wantErr: "recipe \"find-file\" risk is invalid"},
		{name: "invalid execution", mutate: func(recipe *models.Recipe) { recipe.Execution = models.ExecutionType("invalid") }, wantErr: "recipe \"find-file\" execution is invalid"},
		{name: "missing title", mutate: func(recipe *models.Recipe) { recipe.Title = nil }, wantErr: "recipe \"find-file\" title is required"},
		{name: "missing binary", mutate: func(recipe *models.Recipe) { recipe.Binary = "" }, wantErr: "recipe \"find-file\" binary is required for direct execution"},
		{name: "missing shell command", mutate: func(recipe *models.Recipe) {
			recipe.Execution = models.ExecutionTypeShell
			recipe.Binary = ""
			recipe.Command = ""
		}, wantErr: "recipe \"find-file\" command is required for shell execution"},
		{name: "invalid field", mutate: func(recipe *models.Recipe) { recipe.Fields = []models.Field{{Type: models.FieldTypeString}} }, wantErr: "recipe \"find-file\" has invalid field: field name is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recipe := validRecipe
			tt.mutate(&recipe)
			require.EqualError(t, recipe.Validate(), tt.wantErr)
		})
	}
}

// TestExecutionParsing validates execution and risk parsing helpers.
func TestExecutionParsing(t *testing.T) {
	t.Parallel()

	executionType, err := models.ParseExecutionType("direct")
	require.NoError(t, err)
	assert.Equal(t, models.ExecutionTypeDirect, executionType)

	executionType, err = models.ParseExecutionType("shell")
	require.NoError(t, err)
	assert.Equal(t, models.ExecutionTypeShell, executionType)

	_, err = models.ParseExecutionType("missing")
	assert.EqualError(t, err, "unknown execution type \"missing\"")
	assert.False(t, models.ExecutionType("missing").Valid())

	riskLevel, err := models.ParseRiskLevel("safe")
	require.NoError(t, err)
	assert.Equal(t, models.RiskSafe, riskLevel)

	riskLevel, err = models.ParseRiskLevel("elevated")
	require.NoError(t, err)
	assert.Equal(t, models.RiskElevated, riskLevel)

	riskLevel, err = models.ParseRiskLevel("dangerous")
	require.NoError(t, err)
	assert.Equal(t, models.RiskDangerous, riskLevel)

	_, err = models.ParseRiskLevel("missing")
	assert.EqualError(t, err, "unknown risk level \"missing\"")
	assert.False(t, models.RiskLevel("missing").Valid())
}
