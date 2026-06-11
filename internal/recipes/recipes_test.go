package recipes_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	linuxhelper "linux-helper"
	"linux-helper/internal/recipes"
)

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
	assert.GreaterOrEqual(t, len(loaded), 60)

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
	assert.Len(t, categoryCounts, 6)
	assert.GreaterOrEqual(t, categoryCounts["filesystem"], 10)
	assert.GreaterOrEqual(t, categoryCounts["environment"], 10)
	assert.GreaterOrEqual(t, categoryCounts["network"], 10)
	assert.GreaterOrEqual(t, categoryCounts["system"], 10)
	assert.GreaterOrEqual(t, categoryCounts["text"], 10)
	assert.GreaterOrEqual(t, categoryCounts["users"], 10)
}
