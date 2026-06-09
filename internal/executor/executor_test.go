package executor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"linux-helper/internal/executor"
	"linux-helper/internal/models"
)

type fakeRunner struct {
	result  models.ExecutionResult
	err     error
	command string
	args    []string
	shell   string
}

// Run records a direct command invocation.
func (r *fakeRunner) Run(_ context.Context, name string, args ...string) (models.ExecutionResult, error) {
	r.command = name
	r.args = args
	return r.result, r.err
}

// RunShell records a shell command invocation.
func (r *fakeRunner) RunShell(_ context.Context, command string) (models.ExecutionResult, error) {
	r.shell = command
	return r.result, r.err
}

// TestExecuteDirect renders placeholders into arguments.
func TestExecuteDirect(t *testing.T) {
	runner := &fakeRunner{result: models.ExecutionResult{Command: "find ."}}
	result, err := executor.ExecuteDirect(context.Background(), runner, models.Recipe{
		ID:        "find-file",
		Binary:    "find",
		Execution: models.ExecutionTypeDirect,
		Args:      []string{"{{path}}", "-name", "{{filename}}"},
	}, map[string]string{"path": ".", "filename": "*.log"})

	require.NoError(t, err)
	assert.Equal(t, "find", runner.command)
	assert.Equal(t, []string{".", "-name", "*.log"}, runner.args)
	assert.Equal(t, "find .", result.Command)
}

// TestConfirmRisk rejects dangerous recipes without confirmation.
func TestConfirmRisk(t *testing.T) {
	err := executor.ConfirmRisk(models.RiskDangerous, false)
	assert.ErrorIs(t, err, executor.ErrConfirmationRequired)
	assert.NoError(t, executor.ConfirmRisk(models.RiskSafe, false))
}

// TestExecuteShell quotes shell values.
func TestExecuteShell(t *testing.T) {
	runner := &fakeRunner{result: models.ExecutionResult{Command: "echo"}, err: errors.New("boom")}
	_, err := executor.ExecuteShell(context.Background(), runner, models.Recipe{
		ID:        "echo",
		Execution: models.ExecutionTypeShell,
		Command:   "echo {{value}}",
	}, map[string]string{"value": "a b"})

	assert.Error(t, err)
	assert.Equal(t, "echo 'a b'", runner.shell)
}
