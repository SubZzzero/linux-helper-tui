package search

import (
	"strings"

	"linux-helper/internal/models"
)

// Entry stores the text corpus for one recipe.
type Entry struct {
	Recipe models.Recipe
	Text   string
}

// BuildIndex creates searchable entries from recipes.
func BuildIndex(recipes []models.Recipe) []Entry {
	entries := make([]Entry, 0, len(recipes))
	for _, recipe := range recipes {
		parts := []string{recipe.ID, string(recipe.Category), recipe.Title.Resolve("en"), recipe.Description.Resolve("en")}
		parts = append(parts, recipe.Tags...)
		entries = append(entries, Entry{Recipe: recipe, Text: strings.Join(parts, " ")})
	}

	return entries
}
