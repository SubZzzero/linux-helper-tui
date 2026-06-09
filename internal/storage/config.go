package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config stores user preferences.
type Config struct {
	Locale string `yaml:"locale"`
	Theme  string `yaml:"theme"`
}

// Paths describes the standard file layout.
type Paths struct {
	ConfigFile    string
	FavoritesFile string
	RecentFile    string
	RecipesDir    string
	LogFile       string
}

// DefaultConfig returns the starter user configuration.
func DefaultConfig() Config {
	return Config{Locale: "en", Theme: "dark"}
}

// DefaultPaths builds XDG-style paths from a home directory.
func DefaultPaths(home string) Paths {
	configRoot := filepath.Join(home, ".config", "linux-helper")
	dataRoot := filepath.Join(home, ".local", "share", "linux-helper")

	return Paths{
		ConfigFile:    filepath.Join(configRoot, "config.yaml"),
		FavoritesFile: filepath.Join(configRoot, "favorites.yaml"),
		RecentFile:    filepath.Join(configRoot, "recent.yaml"),
		RecipesDir:    filepath.Join(configRoot, "recipes"),
		LogFile:       filepath.Join(dataRoot, "app.log"),
	}
}

// LoadConfig reads the config file or returns defaults when missing.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}

		return Config{}, fmt.Errorf("read config: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	if config.Locale == "" {
		config.Locale = "en"
	}

	if config.Theme == "" {
		config.Theme = "dark"
	}

	return config, nil
}

// SaveConfig writes the config file to disk.
func SaveConfig(path string, config Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
