package models

import "fmt"

// FieldType identifies the input widget and validation strategy.
type FieldType string

const (
	// FieldTypeString is the default text field.
	FieldTypeString FieldType = "string"
)

// LocalizedText stores translated strings by locale key.
type LocalizedText map[string]string

// Resolve returns the value for the requested locale or a fallback.
func (l LocalizedText) Resolve(locale string) string {
	if value, ok := l[locale]; ok && value != "" {
		return value
	}

	if value, ok := l["en"]; ok && value != "" {
		return value
	}

	for _, value := range l {
		if value != "" {
			return value
		}
	}

	return ""
}

// Field describes a recipe input parameter.
type Field struct {
	Name        string        `yaml:"name"`
	Type        FieldType     `yaml:"type"`
	Required    bool          `yaml:"required"`
	Default     string        `yaml:"default"`
	Description LocalizedText `yaml:"description"`
}

// Valid reports whether the field definition is usable.
func (f Field) Valid() error {
	if f.Name == "" {
		return fmt.Errorf("field name is required")
	}

	if f.Type == "" {
		return fmt.Errorf("field %q type is required", f.Name)
	}

	if f.Type != FieldTypeString {
		return fmt.Errorf("field %q type %q is not supported", f.Name, f.Type)
	}

	return nil
}
