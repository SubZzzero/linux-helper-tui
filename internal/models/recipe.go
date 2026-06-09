package models

import "fmt"

// Example shows one concrete recipe invocation.
type Example struct {
	Args        map[string]string `yaml:"args"`
	Description LocalizedText     `yaml:"description"`
}

// Recipe is the primary command definition loaded from YAML.
type Recipe struct {
	ID          string        `yaml:"id"`
	Version     int           `yaml:"version"`
	Type        string        `yaml:"type"`
	Category    Category      `yaml:"category"`
	Risk        RiskLevel     `yaml:"risk"`
	Execution   ExecutionType `yaml:"execution"`
	Binary      string        `yaml:"binary"`
	Command     string        `yaml:"command"`
	Title       LocalizedText `yaml:"title"`
	Description LocalizedText `yaml:"description"`
	Args        []string      `yaml:"args"`
	Fields      []Field       `yaml:"fields"`
	Tags        []string      `yaml:"tags"`
	Examples    []Example     `yaml:"examples"`
}

// Validate checks the recipe for required schema fields.
func (r Recipe) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("recipe id is required")
	}

	if r.Version <= 0 {
		return fmt.Errorf("recipe %q version must be positive", r.ID)
	}

	if r.Type != "recipe" {
		return fmt.Errorf("recipe %q type must be recipe", r.ID)
	}

	if !r.Category.Valid() {
		return fmt.Errorf("recipe %q category is invalid", r.ID)
	}

	if !r.Risk.Valid() {
		return fmt.Errorf("recipe %q risk is invalid", r.ID)
	}

	if !r.Execution.Valid() {
		return fmt.Errorf("recipe %q execution is invalid", r.ID)
	}

	if len(r.Title) == 0 {
		return fmt.Errorf("recipe %q title is required", r.ID)
	}

	if r.Execution == ExecutionTypeDirect && r.Binary == "" {
		return fmt.Errorf("recipe %q binary is required for direct execution", r.ID)
	}

	if r.Execution == ExecutionTypeShell && r.Command == "" {
		return fmt.Errorf("recipe %q command is required for shell execution", r.ID)
	}

	for _, field := range r.Fields {
		if err := field.Valid(); err != nil {
			return fmt.Errorf("recipe %q has invalid field: %w", r.ID, err)
		}
	}

	return nil
}
