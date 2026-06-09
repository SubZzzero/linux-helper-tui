package theme_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"linux-helper/internal/tui/theme"
)

// TestLoadDefinitions reads embedded YAML themes.
func TestLoadDefinitions(t *testing.T) {
	definitions, err := theme.LoadDefinitions(fstest.MapFS{
		"themes/dark.yaml": {Data: []byte("name: dark\nborder_color: \"63\"\naccent_color: \"213\"\n")},
	}, "themes")

	require.NoError(t, err)
	require.Len(t, definitions, 1)
	assert.Equal(t, "63", definitions["dark"].BorderColor)

	definition, err := theme.ResolveDefinition(definitions, "dark")
	require.NoError(t, err)
	assert.Equal(t, "dark", definition.Name)
	assert.NotEmpty(t, theme.NewStyles(definition).Frame)
}

// TestResolveDefinitionMissing returns a useful error.
func TestResolveDefinitionMissing(t *testing.T) {
	_, err := theme.ResolveDefinition(map[string]theme.Definition{}, "missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}
