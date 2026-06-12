package models

import "fmt"

// Category identifies a recipe group.
type Category string

const (
	// CategoryFilesystem groups filesystem commands.
	CategoryFilesystem Category = "filesystem"
	// CategoryEnvironment groups environment inspection commands.
	CategoryEnvironment Category = "environment"
	// CategoryLogs groups log inspection commands.
	CategoryLogs Category = "logs"
	// CategoryNetwork groups networking commands.
	CategoryNetwork Category = "network"
	// CategoryPackages groups package manager inspection commands.
	CategoryPackages Category = "packages"
	// CategoryProcesses groups process inspection commands.
	CategoryProcesses Category = "processes"
	// CategoryServices groups service management inspection commands.
	CategoryServices Category = "services"
	// CategorySystem groups system inspection commands.
	CategorySystem Category = "system"
	// CategoryText groups text-processing commands.
	CategoryText Category = "text"
	// CategoryUsers groups user and session commands.
	CategoryUsers Category = "users"
)

// Valid reports whether the category is known.
func (c Category) Valid() bool {
	switch c {
	case CategoryFilesystem, CategoryEnvironment, CategoryLogs, CategoryNetwork, CategoryPackages, CategoryProcesses, CategoryServices, CategorySystem, CategoryText, CategoryUsers:
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
	case CategoryEnvironment:
		return "Environment"
	case CategoryLogs:
		return "Logs"
	case CategoryNetwork:
		return "Network"
	case CategoryPackages:
		return "Packages"
	case CategoryProcesses:
		return "Processes"
	case CategoryServices:
		return "Services"
	case CategorySystem:
		return "System"
	case CategoryText:
		return "Text"
	case CategoryUsers:
		return "Users"
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
