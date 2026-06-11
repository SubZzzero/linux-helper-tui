package services

import "linux-helper/internal/storage"

// FavoritesService persists and toggles recipe favorites.
type FavoritesService struct {
	store *storage.FavoritesStore
}

// NewFavoritesService builds a favorites service.
func NewFavoritesService(store *storage.FavoritesStore) *FavoritesService {
	return &FavoritesService{store: store}
}

// Load returns the persisted favorite recipe identifiers.
func (s *FavoritesService) Load() ([]string, error) {
	if s == nil || s.store == nil {
		return []string{}, nil
	}

	return s.store.Load()
}

// Toggle flips the favorite state for one recipe identifier.
func (s *FavoritesService) Toggle(recipeID string) (bool, error) {
	if s == nil || s.store == nil {
		return false, nil
	}

	favorites, err := s.store.Load()
	if err != nil {
		return false, err
	}

	updated := make([]string, 0, len(favorites))
	for _, favoriteID := range favorites {
		if favoriteID == recipeID {
			for _, existingID := range favorites {
				if existingID != recipeID {
					updated = append(updated, existingID)
				}
			}

			return false, s.store.Save(updated)
		}
	}

	updated = append([]string{recipeID}, favorites...)
	return true, s.store.Save(updated)
}
