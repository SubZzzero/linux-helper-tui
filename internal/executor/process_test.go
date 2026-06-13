package executor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeExitError struct {
	code int
}

// Error satisfies the error interface for exit code tests.
func (e fakeExitError) Error() string {
	return "exit"
}

// ExitCode returns the synthetic process status.
func (e fakeExitError) ExitCode() int {
	return e.code
}

// TestJoinArgs covers empty and multi-argument rendering.
func TestJoinArgs(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "", joinArgs(nil))
	assert.Equal(t, "one", joinArgs([]string{"one"}))
	assert.Equal(t, "one two three", joinArgs([]string{"one", "two", "three"}))
}

// TestExitCode resolves success, process, and generic failures.
func TestExitCode(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, exitCode(nil))
	assert.Equal(t, -1, exitCode(errors.New("boom")))

	assert.Equal(t, 7, exitCode(fakeExitError{code: 7}))
}

// TestOSRunnerRun captures process output and the rendered command line.
func TestOSRunnerRun(t *testing.T) {
	t.Parallel()
	originalRunProcess := runProcess
	t.Cleanup(func() { runProcess = originalRunProcess })
	runProcess = func(_ context.Context, name string, args ...string) (string, string, error) {
		require.Equal(t, "bash", name)
		assert.Equal(t, []string{"-c", "printf out"}, args)
		return "out", "err", nil
	}

	result, err := OSRunner{}.Run(context.Background(), "bash", "-c", "printf out")
	require.NoError(t, err)
	assert.Equal(t, "bash -c printf out", result.Command)
	assert.Equal(t, "out", result.Stdout)
	assert.Equal(t, "err", result.Stderr)
	assert.Equal(t, 0, result.ExitCode)
}

// TestOSRunnerRunShellPreservesShellCommandOnFailure keeps the user command visible.
func TestOSRunnerRunShellPreservesShellCommandOnFailure(t *testing.T) {
	t.Parallel()
	originalRunProcess := runProcess
	t.Cleanup(func() { runProcess = originalRunProcess })
	runProcess = func(_ context.Context, name string, args ...string) (string, string, error) {
		require.Equal(t, "bash", name)
		assert.Equal(t, []string{"-c", "printf out && printf err >&2 && exit 3"}, args)
		return "out", "err", fakeExitError{code: 3}
	}

	result, err := OSRunner{}.RunShell(context.Background(), "printf out && printf err >&2 && exit 3")
	require.Error(t, err)
	assert.EqualError(t, err, "run shell command: run command: exit")
	assert.Equal(t, "printf out && printf err >&2 && exit 3", result.Command)
	assert.Equal(t, "out", result.Stdout)
	assert.Equal(t, "err", result.Stderr)
	assert.Equal(t, 3, result.ExitCode)
}
