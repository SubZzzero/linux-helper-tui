package storage_test

import (
	"fmt"
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

// TestLoadConfigReturnsDefaultsWhenFileMissing falls back to starter values.
func TestLoadConfigReturnsDefaultsWhenFileMissing(t *testing.T) {
	config, err := storage.LoadConfig(filepath.Join(t.TempDir(), "missing.yaml"))
	require.NoError(t, err)
	assert.Equal(t, storage.DefaultConfig(), config)
}

// TestLoadConfigBackfillsEmptyLocaleAndTheme restores missing config values.
func TestLoadConfigBackfillsEmptyLocaleAndTheme(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, storage.SaveConfig(path, storage.Config{}))

	config, err := storage.LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, storage.Config{Locale: "en", Theme: "dark"}, config)
}

// TestLoadConfigDefaultsOnlyMissingField fills only the empty setting.
func TestLoadConfigDefaultsOnlyMissingField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, storage.SaveConfig(path, storage.Config{Locale: "ua"}))

	config, err := storage.LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, storage.Config{Locale: "ua", Theme: "dark"}, config)

	require.NoError(t, storage.SaveConfig(path, storage.Config{Theme: "light"}))
	config, err = storage.LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, storage.Config{Locale: "en", Theme: "light"}, config)
}

// TestSaveConfigCreatesParentDirectory creates nested config paths.
func TestSaveConfigCreatesParentDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config", "config.yaml")
	require.NoError(t, storage.SaveConfig(path, storage.Config{Locale: "ru", Theme: "light"}))
	config, err := storage.LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, storage.Config{Locale: "ru", Theme: "light"}, config)
}

// TestSaveAndLoadFavorites persists favorite recipe identifiers.
func TestSaveAndLoadFavorites(t *testing.T) {
	store := storage.NewFavoritesStore(filepath.Join(t.TempDir(), "favorites.yaml"))
	require.NoError(t, store.Save([]string{"find-file", "disk-usage"}))

	favorites, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"find-file", "disk-usage"}, favorites)
}

// TestRecentStoreAddPrependsAndTrims keeps the newest commands first.
func TestRecentStoreAddPrependsAndTrims(t *testing.T) {
	store := storage.NewRecentStore(filepath.Join(t.TempDir(), "recent.yaml"))
	for index := 0; index < 12; index++ {
		require.NoError(t, store.Add(fmt.Sprintf("cmd-%02d", index)))
	}

	recent, err := store.Load()
	require.NoError(t, err)
	assert.Len(t, recent, 10)
	assert.Equal(t, "cmd-11", recent[0])
	assert.Equal(t, "cmd-02", recent[9])
}
