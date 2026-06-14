package executor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

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

// TestOSRunnerRunStreaming emits live output and preserves the final buffers.
func TestOSRunnerRunStreaming(t *testing.T) {
	originalRunProcessStreaming := runProcessStreaming
	t.Cleanup(func() { runProcessStreaming = originalRunProcessStreaming })
	runProcessStreaming = func(_ context.Context, name string, sink OutputSink, args ...string) (string, string, error) {
		require.Equal(t, "tail", name)
		assert.Equal(t, []string{"-F", "/var/log/syslog"}, args)
		sink("stdout", "line one\n")
		sink("stderr", "warn\n")
		return "line one\n", "warn\n", nil
	}

	chunks := []string{}
	result, err := OSRunner{}.RunStreaming(context.Background(), "tail", func(stream string, chunk string) {
		chunks = append(chunks, stream+":"+chunk)
	}, "-F", "/var/log/syslog")
	require.NoError(t, err)
	assert.Equal(t, []string{"stdout:line one\n", "stderr:warn\n"}, chunks)
	assert.Equal(t, "tail -F /var/log/syslog", result.Command)
	assert.Equal(t, "line one\n", result.Stdout)
	assert.Equal(t, "warn\n", result.Stderr)
	assert.Equal(t, 0, result.ExitCode)
}

// TestOSRunnerRunStreamingWithTailFollow verifies live chunks from a real tail -F process.
func TestOSRunnerRunStreamingWithTailFollow(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("tail follow test requires linux")
	}

	logPath := filepath.Join(t.TempDir(), "follow.log")
	require.NoError(t, os.WriteFile(logPath, []byte("start\n"), 0o600))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chunks := make(chan string, 8)
	resultCh := make(chan error, 1)
	go func() {
		_, err := OSRunner{}.RunStreaming(ctx, "tail", func(stream string, chunk string) {
			if stream == "stdout" {
				chunks <- chunk
			}
		}, "-n", "1", "-F", logPath)
		resultCh <- err
	}()

	select {
	case chunk := <-chunks:
		assert.Contains(t, chunk, "start")
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial tail output")
	}

	require.NoError(t, os.WriteFile(logPath, []byte("start\nlive\n"), 0o600))
	select {
	case chunk := <-chunks:
		assert.Contains(t, chunk, "live")
	case <-time.After(2 * time.Second):
		t.Fatal("expected streamed tail output")
	}

	cancel()
	select {
	case err := <-resultCh:
		assert.Error(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("expected tail process to stop after cancellation")
	}
}
