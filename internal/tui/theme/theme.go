package theme

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// Definition stores one embedded theme configuration.
type Definition struct {
	Name        string `yaml:"name"`
	BorderColor string `yaml:"border_color"`
	AccentColor string `yaml:"accent_color"`
}

// Styles holds reusable Lip Gloss styles for all screens.
type Styles struct {
	Frame    lipgloss.Style
	Title    lipgloss.Style
	Accent   lipgloss.Style
	Selected lipgloss.Style
	Muted    lipgloss.Style
	Error    lipgloss.Style
	Success  lipgloss.Style
}

// LoadDefinitions reads all YAML theme files from one filesystem.
func LoadDefinitions(fsys fs.FS, basePath string) (map[string]Definition, error) {
	definitions := make(map[string]Definition)
	err := fs.WalkDir(fsys, basePath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("read theme %q: %w", path, err)
		}

		var definition Definition
		if err := yaml.Unmarshal(data, &definition); err != nil {
			return fmt.Errorf("decode theme %q: %w", path, err)
		}

		if definition.Name == "" {
			return fmt.Errorf("decode theme %q: missing theme name", path)
		}

		definitions[definition.Name] = definition
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk themes: %w", err)
	}

	return definitions, nil
}

// Names returns the available theme names in deterministic order.
func Names(definitions map[string]Definition) []string {
	names := make([]string, 0, len(definitions))
	for name := range definitions {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// ResolveDefinition returns one theme by configured name.
func ResolveDefinition(definitions map[string]Definition, name string) (Definition, error) {
	definition, ok := definitions[name]
	if !ok {
		return Definition{}, fmt.Errorf("theme %q not found", name)
	}

	return definition, nil
}

// NewStyles converts one theme definition into reusable screen styles.
func NewStyles(definition Definition) Styles {
	borderColor := lipgloss.Color(definition.BorderColor)
	accentColor := lipgloss.Color(definition.AccentColor)

	return Styles{
		Frame: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(borderColor).
			Padding(1),
		Title: lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true),
		Accent: lipgloss.NewStyle().Foreground(accentColor),
		Selected: lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true),
		Muted:   lipgloss.NewStyle().Faint(true),
		Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true),
		Success: lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true),
	}
}
