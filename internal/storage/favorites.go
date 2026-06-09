package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FavoritesStore persists recipe favorites.
type FavoritesStore struct {
	path string
}

// NewFavoritesStore builds a favorites store.
func NewFavoritesStore(path string) *FavoritesStore {
	return &FavoritesStore{path: path}
}

// Load returns favorite recipe identifiers.
func (s *FavoritesStore) Load() ([]string, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}

		return nil, fmt.Errorf("read favorites: %w", err)
	}

	var favorites []string
	if err := yaml.Unmarshal(data, &favorites); err != nil {
		return nil, fmt.Errorf("decode favorites: %w", err)
	}

	return favorites, nil
}

// Save writes favorite recipe identifiers.
func (s *FavoritesStore) Save(favorites []string) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create favorites dir: %w", err)
	}

	data, err := yaml.Marshal(favorites)
	if err != nil {
		return fmt.Errorf("encode favorites: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write favorites: %w", err)
	}

	return nil
}
