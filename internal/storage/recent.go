package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const maxRecentItems = 10

// RecentStore persists recently executed commands.
type RecentStore struct {
	path string
}

// NewRecentStore builds a recent-command store.
func NewRecentStore(path string) *RecentStore {
	return &RecentStore{path: path}
}

// Load returns the recent command list.
func (s *RecentStore) Load() ([]string, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}

		return nil, fmt.Errorf("read recent: %w", err)
	}

	var recent []string
	if err := yaml.Unmarshal(data, &recent); err != nil {
		return nil, fmt.Errorf("decode recent: %w", err)
	}

	return recent, nil
}

// Add prepends a command and trims the list.
func (s *RecentStore) Add(command string) error {
	recent, err := s.Load()
	if err != nil {
		return err
	}

	updated := append([]string{command}, recent...)
	if len(updated) > maxRecentItems {
		updated = updated[:maxRecentItems]
	}

	return s.Save(updated)
}

// Save writes the recent command list.
func (s *RecentStore) Save(recent []string) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create recent dir: %w", err)
	}

	data, err := yaml.Marshal(recent)
	if err != nil {
		return fmt.Errorf("encode recent: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write recent: %w", err)
	}

	return nil
}
