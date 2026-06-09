package recipes

import (
	"errors"
	"fmt"

	"linux-helper/internal/models"
)

// ErrDuplicateRecipeID reports a duplicate recipe identifier.
var ErrDuplicateRecipeID = errors.New("duplicate recipe id")

// Validate checks a recipe against the expected schema.
func Validate(recipe models.Recipe) error {
	if err := recipe.Validate(); err != nil {
		return fmt.Errorf("validate recipe: %w", err)
	}

	return nil
}
