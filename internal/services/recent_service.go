package services

import "linux-helper/internal/storage"

// RecentService loads persisted recently executed commands.
type RecentService struct {
	store *storage.RecentStore
}

// NewRecentService builds a recent-command service.
func NewRecentService(store *storage.RecentStore) *RecentService {
	return &RecentService{store: store}
}

// Load returns the persisted recent command list.
func (s *RecentService) Load() ([]string, error) {
	if s == nil || s.store == nil {
		return []string{}, nil
	}

	return s.store.Load()
}
