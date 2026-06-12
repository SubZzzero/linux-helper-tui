package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"linux-helper/internal/app"
)

// TestBootstrapUsesDefaultsWhenConfigMissing starts with embedded English defaults.
func TestBootstrapUsesDefaultsWhenConfigMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	model, closeLog, err := app.Bootstrap()
	require.NoError(t, err)
	defer func() { require.NoError(t, closeLog()) }()

	view := model.View()
	assert.Contains(t, view, "Recent commands")
	assert.Contains(t, view, "ctrl+l locale")
	assert.FileExists(t, filepath.Join(home, ".local", "share", "linux-helper", "app.log"))
}

// TestBootstrapUsesPersistedLocaleAndTheme applies the saved locale at startup.
func TestBootstrapUsesPersistedLocaleAndTheme(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configPath := filepath.Join(home, ".config", "linux-helper", "config.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0o755))
	require.NoError(t, os.WriteFile(configPath, []byte("locale: ua\ntheme: light\n"), 0o644))

	model, closeLog, err := app.Bootstrap()
	require.NoError(t, err)
	defer func() { require.NoError(t, closeLog()) }()

	view := model.View()
	assert.Contains(t, view, "Останні команди")
	assert.Contains(t, view, "ctrl+l мова")
}
