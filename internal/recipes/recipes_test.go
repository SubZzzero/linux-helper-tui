package recipes_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
