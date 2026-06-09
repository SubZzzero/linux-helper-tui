package executor

import (
	"context"
	"fmt"
	"strings"

	"linux-helper/internal/models"
)

// ExecuteShell launches a shell-execution recipe.
func ExecuteShell(ctx context.Context, runner CommandRunner, recipe models.Recipe, values map[string]string) (models.ExecutionResult, error) {
	command := recipe.Command
	for key, value := range values {
		command = strings.ReplaceAll(command, "{{"+key+"}}", shellQuote(value))
	}

	if strings.Contains(command, "{{") {
		return models.ExecutionResult{}, fmt.Errorf("unresolved shell template %q", recipe.Command)
	}

	result, err := runner.RunShell(ctx, command)
	if err != nil {
		return result, fmt.Errorf("execute shell recipe %q: %w", recipe.ID, err)
	}

	return result, nil
}

// shellQuote escapes one argument for bash -c execution.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `"'"'`) + "'"
}
