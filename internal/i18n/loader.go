package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"

	"linux-helper/internal/models"
)

// LoadLocales reads all JSON locale files from one filesystem.
func LoadLocales(fsys fs.FS, basePath string) (map[string]models.LocalizedText, error) {
	result := make(map[string]models.LocalizedText)
	err := fs.WalkDir(fsys, basePath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("read locale %q: %w", path, err)
		}

		catalogue := make(map[string]string)
		if err := json.Unmarshal(data, &catalogue); err != nil {
			return fmt.Errorf("decode locale %q: %w", path, err)
		}

		locale := filepath.Base(path[:len(path)-len(filepath.Ext(path))])
		for key, value := range catalogue {
			text := result[key]
			if text == nil {
				text = models.LocalizedText{}
			}
			text[locale] = value
			result[key] = text
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk locales: %w", err)
	}

	return result, nil
}
