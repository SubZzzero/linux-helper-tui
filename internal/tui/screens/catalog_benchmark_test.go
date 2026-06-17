package screens_test

import (
	"fmt"
	"testing"

	"linux-helper/internal/models"
	"linux-helper/internal/tui/screens"
)

func benchmarkCatalogRecipes(count int) []models.Recipe {
	recipes := make([]models.Recipe, 0, count)
	categories := []models.Category{
		models.CategoryDocker,
		models.CategoryFilesystem,
		models.CategoryEnvironment,
		models.CategoryLogs,
		models.CategoryNetwork,
		models.CategoryPackages,
		models.CategoryProcesses,
		models.CategoryServices,
		models.CategorySystem,
		models.CategoryText,
		models.CategoryTroubleshooting,
		models.CategoryUsers,
	}

	for index := 0; index < count; index++ {
		category := categories[index%len(categories)]
		recipes = append(recipes, models.Recipe{
			ID:          fmt.Sprintf("recipe-%03d", index),
			Version:     1,
			Type:        "recipe",
			Category:    category,
			Risk:        models.RiskSafe,
			Execution:   models.ExecutionTypeDirect,
			Binary:      "printf",
			Args:        []string{"ok"},
			Title:       models.LocalizedText{"en": fmt.Sprintf("Recipe %03d", index)},
			Description: models.LocalizedText{"en": "Synthetic benchmark recipe"},
		})
	}

	return recipes
}

// BenchmarkNewCatalogModel measures catalog-first discovery initialization at corpus scale.
func BenchmarkNewCatalogModel(b *testing.B) {
	recipes := benchmarkCatalogRecipes(200)
	styles := testStyles()
	b.ReportAllocs()

	for range b.N {
		model := screens.NewCatalogModel(recipes, "en", styles, nil, nil, "linux-helper", "Empty", "Recent", "No recent commands yet.", "Catalog help")
		if view := model.View(); view == "" {
			b.Fatal("expected non-empty catalog view")
		}
	}
}

// BenchmarkCatalogSetSelectedCategory measures category filtering on a larger corpus.
func BenchmarkCatalogSetSelectedCategory(b *testing.B) {
	recipes := benchmarkCatalogRecipes(200)
	styles := testStyles()
	model := screens.NewCatalogModel(recipes, "en", styles, nil, nil, "linux-helper", "Empty", "Recent", "No recent commands yet.", "Catalog help")
	categories := []models.Category{
		models.CategoryDocker,
		models.CategoryFilesystem,
		models.CategoryEnvironment,
		models.CategoryLogs,
		models.CategoryNetwork,
		models.CategoryPackages,
		models.CategoryProcesses,
		models.CategoryServices,
		models.CategorySystem,
		models.CategoryText,
		models.CategoryTroubleshooting,
		models.CategoryUsers,
	}
	b.ReportAllocs()

	for index := 0; index < b.N; index++ {
		model.SetSelectedCategory(categories[index%len(categories)])
		if view := model.View(); view == "" {
			b.Fatal("expected non-empty category view")
		}
	}
}
