package recipes_test

import (
	"io/fs"
	"regexp"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	linuxhelper "linux-helper"
	"linux-helper/internal/models"
	"linux-helper/internal/recipes"
)

var placeholderPattern = regexp.MustCompile(`\{\{([a-zA-Z0-9_-]+)\}\}`)

// TestLoaderLoad reads and validates YAML recipes.
func TestLoaderLoad(t *testing.T) {
	fsys := fstest.MapFS{
		"recipes/find.yaml": {Data: []byte(`id: find-file
version: 1
type: recipe
category: filesystem
risk: safe
execution: direct
binary: find
title:
  en: Find file
description:
  en: Search files
args: ["{{path}}"]
fields:
  - name: path
    type: string
`)},
	}

	loader := recipes.NewLoader(fsys, nil, "recipes")
	loaded, err := loader.Load()
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	assert.Equal(t, "find-file", loaded[0].ID)
	assert.NoError(t, recipes.Validate(loaded[0]))
}

// TestRegistryGet returns recipes by identifier.
func TestRegistryGet(t *testing.T) {
	loader := recipes.NewLoader(fstest.MapFS{
		"recipes/find.yaml": {Data: []byte(`id: find-file
version: 1
type: recipe
category: filesystem
risk: safe
execution: direct
binary: find
title:
  en: Find file
description:
  en: Search files
`)},
	}, nil, "recipes")

	loaded, err := loader.Load()
	require.NoError(t, err)

	registry, err := recipes.NewRegistry(loaded)
	require.NoError(t, err)

	recipe, err := registry.Get("find-file")
	require.NoError(t, err)
	assert.Equal(t, "find-file", recipe.ID)
}

// TestEmbeddedRecipeCorpusLoads validates all bundled recipe files.
func TestEmbeddedRecipeCorpusLoads(t *testing.T) {
	recipeFS, err := fs.Sub(linuxhelper.Assets, "assets/recipes")
	require.NoError(t, err)

	loader := recipes.NewLoader(recipeFS, nil, ".")
	loaded, err := loader.Load()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(loaded), 100)

	ids := make(map[string]struct{}, len(loaded))
	categoryCounts := make(map[string]int)
	for _, recipe := range loaded {
		ids[recipe.ID] = struct{}{}
		categoryCounts[string(recipe.Category)]++
		assert.NoError(t, recipes.Validate(recipe))
	}

	assert.Contains(t, ids, "find-file")
	assert.Contains(t, ids, "ping-host")
	assert.Contains(t, ids, "grep-pattern")
	assert.Contains(t, ids, "list-directory")
	assert.Contains(t, ids, "current-user")
	assert.Contains(t, ids, "print-environment")
	assert.Contains(t, ids, "top-cpu-processes")
	assert.Contains(t, ids, "directory-disk-usage")
	assert.Contains(t, ids, "locale-settings")
	assert.Contains(t, ids, "route-table")
	assert.Contains(t, ids, "mounted-filesystems")
	assert.Contains(t, ids, "diff-files")
	assert.Contains(t, ids, "passwd-entry")
	assert.Contains(t, ids, "socket-summary")
	assert.Contains(t, ids, "list-locales")
	assert.Contains(t, ids, "comm-files")
	assert.Contains(t, ids, "group-entry")
	assert.Contains(t, ids, "checksum-file")
	assert.Contains(t, ids, "block-devices")
	assert.Contains(t, ids, "show-umask")
	assert.Contains(t, ids, "interface-summary")
	assert.Contains(t, ids, "number-lines")
	assert.Contains(t, ids, "user-login-shell")
	assert.Contains(t, ids, "tail-log-file")
	assert.Contains(t, ids, "list-installed-packages")
	assert.Contains(t, ids, "process-tree")
	assert.Contains(t, ids, "service-status")
	assert.Contains(t, ids, "journal-since")
	assert.Contains(t, ids, "search-installed-packages")
	assert.Contains(t, ids, "top-memory-processes")
	assert.Contains(t, ids, "running-services")
	assert.Contains(t, ids, "journal-priority-errors")
	assert.Contains(t, ids, "port-owner")
	assert.Contains(t, ids, "systemd-critical-chain")
	assert.Len(t, categoryCounts, 11)
	assert.GreaterOrEqual(t, categoryCounts["filesystem"], 10)
	assert.GreaterOrEqual(t, categoryCounts["environment"], 10)
	assert.GreaterOrEqual(t, categoryCounts["logs"], 8)
	assert.GreaterOrEqual(t, categoryCounts["network"], 10)
	assert.GreaterOrEqual(t, categoryCounts["packages"], 8)
	assert.GreaterOrEqual(t, categoryCounts["processes"], 8)
	assert.GreaterOrEqual(t, categoryCounts["services"], 8)
	assert.GreaterOrEqual(t, categoryCounts["system"], 10)
	assert.GreaterOrEqual(t, categoryCounts["text"], 10)
	assert.GreaterOrEqual(t, categoryCounts["troubleshooting"], 10)
	assert.GreaterOrEqual(t, categoryCounts["users"], 10)
}

// TestEmbeddedRecipeTemplatePlaceholdersMatchFields keeps recipes and fields aligned.
func TestEmbeddedRecipeTemplatePlaceholdersMatchFields(t *testing.T) {
	recipes := loadEmbeddedRecipes(t)
	for _, recipe := range recipes {
		fieldNames := make(map[string]struct{}, len(recipe.Fields))
		for _, field := range recipe.Fields {
			_, duplicate := fieldNames[field.Name]
			assert.Falsef(t, duplicate, "recipe %q has duplicate field %q", recipe.ID, field.Name)
			fieldNames[field.Name] = struct{}{}
		}

		for _, placeholder := range recipePlaceholders(recipe) {
			_, ok := fieldNames[placeholder]
			assert.Truef(t, ok, "recipe %q placeholder %q is missing a field declaration", recipe.ID, placeholder)
		}
	}
}

// TestEmbeddedRecipeExamplesCoverParameterizedRecipes keeps examples useful for field-driven recipes.
func TestEmbeddedRecipeExamplesCoverParameterizedRecipes(t *testing.T) {
	recipes := loadEmbeddedRecipes(t)
	for _, recipe := range recipes {
		placeholders := recipePlaceholders(recipe)
		if len(recipe.Fields) == 0 && len(placeholders) == 0 {
			continue
		}

		require.NotEmptyf(t, recipe.Examples, "recipe %q should include an example", recipe.ID)
		fieldNames := make(map[string]struct{}, len(recipe.Fields))
		for _, field := range recipe.Fields {
			fieldNames[field.Name] = struct{}{}
		}

		for _, example := range recipe.Examples {
			assert.NotEmptyf(t, example.Description.Resolve("en"), "recipe %q example is missing an English description", recipe.ID)
			for key := range example.Args {
				_, ok := fieldNames[key]
				assert.Truef(t, ok, "recipe %q example references unknown field %q", recipe.ID, key)
			}

			for _, placeholder := range placeholders {
				_, ok := example.Args[placeholder]
				assert.Truef(t, ok, "recipe %q example is missing placeholder field %q", recipe.ID, placeholder)
			}
		}
	}
}

func loadEmbeddedRecipes(t *testing.T) []models.Recipe {
	t.Helper()
	recipeFS, err := fs.Sub(linuxhelper.Assets, "assets/recipes")
	require.NoError(t, err)

	loader := recipes.NewLoader(recipeFS, nil, ".")
	loaded, err := loader.Load()
	require.NoError(t, err)
	return loaded
}

func recipePlaceholders(recipe models.Recipe) []string {
	unique := make(map[string]struct{})
	for _, arg := range recipe.Args {
		for _, match := range placeholderPattern.FindAllStringSubmatch(arg, -1) {
			unique[match[1]] = struct{}{}
		}
	}

	for _, match := range placeholderPattern.FindAllStringSubmatch(recipe.Command, -1) {
		unique[match[1]] = struct{}{}
	}

	placeholders := make([]string, 0, len(unique))
	for name := range unique {
		placeholders = append(placeholders, name)
	}

	return placeholders
}
