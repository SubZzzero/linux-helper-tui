package recipes

import (
	"errors"

	"linux-helper/internal/models"
)

// ErrRecipeNotFound is returned when a recipe is missing from the registry.
var ErrRecipeNotFound = errors.New("recipe not found")

// Registry stores recipes indexed by identifier.
type Registry struct {
	byID map[string]models.Recipe
	all  []models.Recipe
}

// NewRegistry constructs a validated recipe registry.
func NewRegistry(recipes []models.Recipe) (*Registry, error) {
	byID := make(map[string]models.Recipe, len(recipes))
	all := make([]models.Recipe, 0, len(recipes))

	for _, recipe := range recipes {
		if _, exists := byID[recipe.ID]; exists {
			return nil, ErrDuplicateRecipeID
		}

		byID[recipe.ID] = recipe
		all = append(all, recipe)
	}

	return &Registry{byID: byID, all: all}, nil
}

// All returns recipes in load order.
func (r *Registry) All() []models.Recipe {
	recipes := make([]models.Recipe, len(r.all))
	copy(recipes, r.all)

	return recipes
}

// Get returns one recipe by identifier.
func (r *Registry) Get(id string) (models.Recipe, error) {
	recipe, ok := r.byID[id]
	if !ok {
		return models.Recipe{}, ErrRecipeNotFound
	}

	return recipe, nil
}
