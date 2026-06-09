package services

import (
	"linux-helper/internal/models"
	"linux-helper/internal/search"
)

// SearchService exposes recipe search over an index.
type SearchService struct {
	index *search.Index
}

// NewSearchService builds a search service.
func NewSearchService(recipes []models.Recipe) *SearchService {
	return &SearchService{index: search.NewIndex(recipes)}
}

// Search returns recipes matching the query.
func (s *SearchService) Search(query string) ([]models.Recipe, error) {
	return s.index.Search(query), nil
}
