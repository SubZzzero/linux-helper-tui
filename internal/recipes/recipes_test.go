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
	assert.Len(t, loaded, 28)

	ids := make(map[string]struct{}, len(loaded))
	for _, recipe := range loaded {
		ids[recipe.ID] = struct{}{}
		assert.NoError(t, recipes.Validate(recipe))
	}

	assert.Contains(t, ids, "find-file")
	assert.Contains(t, ids, "ping-host")
	assert.Contains(t, ids, "grep-pattern")
	assert.Contains(t, ids, "list-directory")
	assert.Contains(t, ids, "current-user")
	assert.Contains(t, ids, "print-environment")
	assert.Contains(t, ids, "top-cpu-processes")
}
