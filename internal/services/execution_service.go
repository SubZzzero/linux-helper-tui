package services

import (
	"context"
	"fmt"

	"linux-helper/internal/executor"
	"linux-helper/internal/models"
	"linux-helper/internal/storage"
)

// ExecutionService orchestrates recipe execution and recent storage.
type ExecutionService struct {
	runner executor.CommandRunner
	recent *storage.RecentStore
}

// NewExecutionService builds an execution service.
func NewExecutionService(runner executor.CommandRunner, recent *storage.RecentStore) *ExecutionService {
	return &ExecutionService{runner: runner, recent: recent}
}

// Execute runs a recipe and records it when successful.
func (s *ExecutionService) Execute(ctx context.Context, recipe models.Recipe, values map[string]string, confirmed bool) (models.ExecutionResult, error) {
	if err := executor.ConfirmRisk(recipe.Risk, confirmed); err != nil {
		return models.ExecutionResult{}, err
	}

	var (
		result models.ExecutionResult
		err    error
	)

	switch recipe.Execution {
	case models.ExecutionTypeDirect:
		result, err = executor.ExecuteDirect(ctx, s.runner, recipe, values)
	case models.ExecutionTypeShell:
		result, err = executor.ExecuteShell(ctx, s.runner, recipe, values)
	default:
		err = fmt.Errorf("unsupported execution type %q", recipe.Execution)
	}
	if err != nil {
		return result, err
	}

	if s.recent != nil {
		if err := s.recent.Add(result.Command); err != nil {
			return result, fmt.Errorf("record recent command: %w", err)
		}
	}

	return result, nil
}
