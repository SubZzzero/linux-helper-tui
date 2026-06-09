package recipes

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	"linux-helper/internal/models"
)

// Loader reads recipes from embedded and optional override filesystems.
type Loader struct {
	embedded fs.FS
	override fs.FS
	basePath string
}

// NewLoader builds a recipe loader.
func NewLoader(embedded fs.FS, override fs.FS, basePath string) *Loader {
	return &Loader{embedded: embedded, override: override, basePath: basePath}
}

// Load reads, parses, and validates recipes.
func (l *Loader) Load() ([]models.Recipe, error) {
	recipes, err := l.loadFromFS(l.embedded)
	if err != nil {
		return nil, err
	}

	if l.override == nil {
		return recipes, nil
	}

	overrides, err := l.loadFromFS(l.override)
	if err != nil {
		return nil, err
	}

	merged := make(map[string]models.Recipe, len(recipes)+len(overrides))
	for _, recipe := range recipes {
		merged[recipe.ID] = recipe
	}

	for _, recipe := range overrides {
		merged[recipe.ID] = recipe
	}

	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]models.Recipe, 0, len(keys))
	for _, key := range keys {
		result = append(result, merged[key])
	}

	return result, nil
}

// loadFromFS reads all YAML recipe files from one filesystem.
func (l *Loader) loadFromFS(fsys fs.FS) ([]models.Recipe, error) {
	entries := make([]string, 0)
	err := fs.WalkDir(fsys, l.basePath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".yaml" {
			entries = append(entries, path)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []models.Recipe{}, nil
		}

		return nil, fmt.Errorf("walk recipes: %w", err)
	}

	sort.Strings(entries)
	recipes := make([]models.Recipe, 0, len(entries))
	for _, entry := range entries {
		data, readErr := fs.ReadFile(fsys, entry)
		if readErr != nil {
			return nil, fmt.Errorf("read recipe %q: %w", entry, readErr)
		}

		recipe, parseErr := Parse(data)
		if parseErr != nil {
			return nil, fmt.Errorf("%s: %w", entry, parseErr)
		}

		if validateErr := Validate(recipe); validateErr != nil {
			return nil, fmt.Errorf("%s: %w", entry, validateErr)
		}

		recipes = append(recipes, recipe)
	}

	return recipes, nil
}
