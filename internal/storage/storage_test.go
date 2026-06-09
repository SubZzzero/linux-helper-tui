package storage_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"linux-helper/internal/storage"
)

// TestDefaultPaths builds the expected XDG-style files.
func TestDefaultPaths(t *testing.T) {
	paths := storage.DefaultPaths("/home/test")
	assert.Equal(t, "/home/test/.config/linux-helper/config.yaml", paths.ConfigFile)
	assert.Equal(t, "/home/test/.local/share/linux-helper/app.log", paths.LogFile)
}

// TestSaveAndLoadConfig persists simple config values.
func TestSaveAndLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, storage.SaveConfig(path, storage.Config{Locale: "ua", Theme: "light"}))

	config, err := storage.LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, "ua", config.Locale)
	assert.Equal(t, "light", config.Theme)
}
