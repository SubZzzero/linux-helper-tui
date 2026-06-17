package recipes_test

import (
	"errors"
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

// TestLoaderLoadOverridesMergeAndSort prefers overrides and returns sorted results.
func TestLoaderLoadOverridesMergeAndSort(t *testing.T) {
	t.Parallel()

	embedded := fstest.MapFS{
		"recipes/z-last.yaml": {Data: []byte(`id: z-last
version: 1
type: recipe
category: filesystem
risk: safe
execution: direct
binary: ls
title:
  en: Last
`)},
		"recipes/shared.yaml": {Data: []byte(`id: shared
version: 1
type: recipe
category: filesystem
risk: safe
execution: direct
binary: find
title:
  en: Embedded
`)},
	}

	override := fstest.MapFS{
		"recipes/a-first.yaml": {Data: []byte(`id: a-first
version: 1
type: recipe
category: filesystem
risk: safe
execution: direct
binary: pwd
title:
  en: First
`)},
		"recipes/shared.yaml": {Data: []byte(`id: shared
version: 1
type: recipe
category: filesystem
risk: safe
execution: direct
binary: printf
title:
  en: Override
`)},
	}

	loaded, err := recipes.NewLoader(embedded, override, "recipes").Load()
	require.NoError(t, err)
	require.Len(t, loaded, 3)
	assert.Equal(t, []string{"a-first", "shared", "z-last"}, []string{loaded[0].ID, loaded[1].ID, loaded[2].ID})
	assert.Equal(t, "printf", loaded[1].Binary)
	assert.Equal(t, "Override", loaded[1].Title.Resolve("en"))
}

// TestLoaderLoadMissingOverridePath ignores an absent override tree.
func TestLoaderLoadMissingOverridePath(t *testing.T) {
	t.Parallel()

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
`)},
	}, fstest.MapFS{}, "recipes")

	loaded, err := loader.Load()
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	assert.Equal(t, "find-file", loaded[0].ID)
}

// TestLoaderLoadErrors wraps parser, validator, and filesystem failures.
func TestLoaderLoadErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		loader  *recipes.Loader
		wantErr string
	}{
		{
			name: "parse error",
			loader: recipes.NewLoader(fstest.MapFS{
				"recipes/bad.yaml": {Data: []byte("id: [")},
			}, nil, "recipes"),
			wantErr: "recipes/bad.yaml: parse recipe yaml:",
		},
		{
			name: "validation error",
			loader: recipes.NewLoader(fstest.MapFS{
				"recipes/invalid.yaml": {Data: []byte(`id: invalid
version: 1
type: recipe
category: filesystem
risk: safe
execution: direct
title:
  en: Invalid
`)},
			}, nil, "recipes"),
			wantErr: "recipes/invalid.yaml: validate recipe: recipe \"invalid\" binary is required for direct execution",
		},
		{
			name:    "walk error",
			loader:  recipes.NewLoader(errFS{err: errors.New("boom")}, nil, "recipes"),
			wantErr: "walk recipes: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tt.loader.Load()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// TestRegistryBranches covers duplicate IDs, missing recipes, and copy semantics.
func TestRegistryBranches(t *testing.T) {
	t.Parallel()

	_, err := recipes.NewRegistry([]models.Recipe{{ID: "dup"}, {ID: "dup"}})
	assert.ErrorIs(t, err, recipes.ErrDuplicateRecipeID)

	registry, err := recipes.NewRegistry([]models.Recipe{{ID: "one"}, {ID: "two"}})
	require.NoError(t, err)

	_, err = registry.Get("missing")
	assert.ErrorIs(t, err, recipes.ErrRecipeNotFound)

	all := registry.All()
	all[0].ID = "changed"
	assert.Equal(t, "one", registry.All()[0].ID)
}

// TestEmbeddedRecipeCorpusLoads validates all bundled recipe files.
func TestEmbeddedRecipeCorpusLoads(t *testing.T) {
	recipeFS, err := fs.Sub(linuxhelper.Assets, "assets/recipes")
	require.NoError(t, err)

	loader := recipes.NewLoader(recipeFS, nil, ".")
	loaded, err := loader.Load()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(loaded), 150)

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
	assert.Contains(t, ids, "path-entries")
	assert.Contains(t, ids, "recent-files")
	assert.Contains(t, ids, "journal-disk-usage")
	assert.Contains(t, ids, "route-to-host")
	assert.Contains(t, ids, "repository-package-search")
	assert.Contains(t, ids, "process-environment")
	assert.Contains(t, ids, "service-timers")
	assert.Contains(t, ids, "cpu-info-summary")
	assert.Contains(t, ids, "unique-line-counts")
	assert.Contains(t, ids, "user-home-disk-usage")
	assert.Contains(t, ids, "memory-pressure")
	assert.Contains(t, ids, "list-broken-symlinks")
	assert.Contains(t, ids, "journal-unit-since")
	assert.Contains(t, ids, "process-parent")
	assert.Contains(t, ids, "service-restart-policy")
	assert.Contains(t, ids, "docker-ps")
	assert.Contains(t, ids, "docker-system-df")
	assert.Len(t, categoryCounts, 12)
	assert.GreaterOrEqual(t, categoryCounts["docker"], 10)
	assert.GreaterOrEqual(t, categoryCounts["filesystem"], 14)
	assert.GreaterOrEqual(t, categoryCounts["environment"], 14)
	assert.GreaterOrEqual(t, categoryCounts["logs"], 12)
	assert.GreaterOrEqual(t, categoryCounts["network"], 14)
	assert.GreaterOrEqual(t, categoryCounts["packages"], 12)
	assert.GreaterOrEqual(t, categoryCounts["processes"], 12)
	assert.GreaterOrEqual(t, categoryCounts["services"], 12)
	assert.GreaterOrEqual(t, categoryCounts["system"], 14)
	assert.GreaterOrEqual(t, categoryCounts["text"], 14)
	assert.GreaterOrEqual(t, categoryCounts["troubleshooting"], 15)
	assert.GreaterOrEqual(t, categoryCounts["users"], 14)
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

type errFS struct {
	err error
}

// Open returns the configured filesystem error.
func (f errFS) Open(string) (fs.File, error) {
	return nil, f.err
}
