package executor

import (
	"context"
	"fmt"
	"strings"

	"linux-helper/internal/models"
)

// ExecuteDirect launches a direct-execution recipe.
func ExecuteDirect(ctx context.Context, runner CommandRunner, recipe models.Recipe, values map[string]string) (models.ExecutionResult, error) {
	args := make([]string, 0, len(recipe.Args))
	for _, arg := range recipe.Args {
		rendered, err := renderTemplate(arg, values)
		if err != nil {
			return models.ExecutionResult{}, err
		}
		args = append(args, rendered)
	}

	result, err := runner.Run(ctx, recipe.Binary, args...)
	if err != nil {
		return result, fmt.Errorf("execute direct recipe %q: %w", recipe.ID, err)
	}

	return result, nil
}

// renderTemplate replaces {{name}} placeholders with field values.
func renderTemplate(template string, values map[string]string) (string, error) {
	result := template
	for key, value := range values {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}

	if strings.Contains(result, "{{") {
		return "", fmt.Errorf("unresolved template %q", template)
	}

	return result, nil
}
