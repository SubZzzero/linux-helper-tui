package search_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"linux-helper/internal/models"
	"linux-helper/internal/search"
)

// TestIndexSearch returns matching recipes in-memory.
func TestIndexSearch(t *testing.T) {
	index := search.NewIndex([]models.Recipe{
		{ID: "find-file", Category: models.CategoryFilesystem, Title: models.LocalizedText{"en": "Find file"}, Description: models.LocalizedText{"en": "Find files"}},
		{ID: "disk-usage", Category: models.CategorySystem, Title: models.LocalizedText{"en": "Disk usage"}, Description: models.LocalizedText{"en": "Show disk usage"}},
	})

	results := index.Search("disk")
	assert.Len(t, results, 1)
	assert.Equal(t, "disk-usage", results[0].ID)
}
