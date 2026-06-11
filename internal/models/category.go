package models

import "fmt"

// Category identifies a recipe group.
type Category string

const (
	// CategoryFilesystem groups filesystem commands.
	CategoryFilesystem Category = "filesystem"
	// CategoryNetwork groups networking commands.
	CategoryNetwork Category = "network"
	// CategorySystem groups system inspection commands.
	CategorySystem Category = "system"
	// CategoryText groups text-processing commands.
	CategoryText Category = "text"
)

// Valid reports whether the category is known.
func (c Category) Valid() bool {
	switch c {
	case CategoryFilesystem, CategoryNetwork, CategorySystem, CategoryText:
		return true
	default:
		return false
	}
}

// DisplayName returns the human-readable category label.
func (c Category) DisplayName() string {
	switch c {
	case CategoryFilesystem:
		return "Filesystem"
	case CategoryNetwork:
		return "Network"
	case CategorySystem:
		return "System"
	case CategoryText:
		return "Text"
	default:
		return string(c)
	}
}

// ParseCategory converts a raw value into a validated category.
func ParseCategory(value string) (Category, error) {
	category := Category(value)
	if !category.Valid() {
		return "", fmt.Errorf("unknown category %q", value)
	}

	return category, nil
}
