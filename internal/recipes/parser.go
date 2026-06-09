package recipes

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"linux-helper/internal/models"
)

// Parse decodes one YAML recipe document.
func Parse(data []byte) (models.Recipe, error) {
	var recipe models.Recipe
	if err := yaml.Unmarshal(data, &recipe); err != nil {
		return models.Recipe{}, fmt.Errorf("parse recipe yaml: %w", err)
	}

	return recipe, nil
}
