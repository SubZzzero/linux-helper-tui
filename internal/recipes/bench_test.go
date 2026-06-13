package recipes_test

import (
	"io/fs"
	"testing"

	linuxhelper "linux-helper"
	"linux-helper/internal/recipes"
	"linux-helper/internal/services"
)

func benchmarkRecipeFS(b *testing.B) fs.FS {
	b.Helper()

	recipeFS, err := fs.Sub(linuxhelper.Assets, "assets/recipes")
	if err != nil {
		b.Fatalf("open embedded recipes: %v", err)
	}

	return recipeFS
}

// BenchmarkLoaderLoadEmbedded measures catalog-first recipe loading from embedded assets.
func BenchmarkLoaderLoadEmbedded(b *testing.B) {
	loader := recipes.NewLoader(benchmarkRecipeFS(b), nil, ".")
	b.ReportAllocs()

	for range b.N {
		loaded, err := loader.Load()
		if err != nil {
			b.Fatalf("load recipes: %v", err)
		}
		if len(loaded) == 0 {
			b.Fatal("expected embedded recipes")
		}
	}
}

// BenchmarkRecipeServiceAllEmbedded measures registry construction and catalog enumeration.
func BenchmarkRecipeServiceAllEmbedded(b *testing.B) {
	b.ReportAllocs()

	for range b.N {
		service, err := services.NewRecipeService(recipes.NewLoader(benchmarkRecipeFS(b), nil, "."))
		if err != nil {
			b.Fatalf("create recipe service: %v", err)
		}

		all := service.All()
		if len(all) == 0 {
			b.Fatal("expected embedded recipes")
		}
	}
}
