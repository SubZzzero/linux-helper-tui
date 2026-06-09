package services

import (
	"fmt"

	"linux-helper/internal/models"
	"linux-helper/internal/recipes"
)

// RecipeLoader reads validated recipes from a source.
type RecipeLoader interface {
	Load() ([]models.Recipe, error)
}

// RecipeService provides registry-backed recipe access.
type RecipeService struct {
	registry *recipes.Registry
}

// NewRecipeService loads recipes and constructs a registry.
func NewRecipeService(loader RecipeLoader) (*RecipeService, error) {
	loaded, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("load recipes: %w", err)
	}

	registry, err := recipes.NewRegistry(loaded)
	if err != nil {
		return nil, fmt.Errorf("create recipe registry: %w", err)
	}

	return &RecipeService{registry: registry}, nil
}

// All returns all known recipes.
func (s *RecipeService) All() []models.Recipe {
	return s.registry.All()
}

// Get returns one recipe by identifier.
func (s *RecipeService) Get(id string) (models.Recipe, error) {
	return s.registry.Get(id)
}
