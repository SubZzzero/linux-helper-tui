package search

import (
	"strings"

	"github.com/sahilm/fuzzy"

	"linux-helper/internal/models"
)

// Index provides in-memory fuzzy search over recipes.
type Index struct {
	entries []Entry
	text    []string
}

// NewIndex constructs a search index from recipes.
func NewIndex(recipes []models.Recipe) *Index {
	entries := BuildIndex(recipes)
	text := make([]string, 0, len(entries))
	for _, entry := range entries {
		text = append(text, entry.Text)
	}

	return &Index{entries: entries, text: text}
}

// Search returns recipes matching the query.
func (i *Index) Search(query string) []models.Recipe {
	if strings.TrimSpace(query) == "" {
		return allRecipes(i.entries)
	}

	matches := fuzzy.Find(query, i.text)
	sortMatches(matches)

	results := make([]models.Recipe, 0, len(matches))
	for _, match := range matches {
		results = append(results, i.entries[match.Index].Recipe)
	}

	return results
}

// allRecipes returns all recipes from index entries.
func allRecipes(entries []Entry) []models.Recipe {
	recipes := make([]models.Recipe, 0, len(entries))
	for _, entry := range entries {
		recipes = append(recipes, entry.Recipe)
	}

	return recipes
}
