package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"linux-helper/internal/models"
)

var runProcess = func(ctx context.Context, name string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

var runProcessStreaming = func(ctx context.Context, name string, sink OutputSink, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdout, outputWriter{stream: "stdout", sink: sink})
	cmd.Stderr = io.MultiWriter(&stderr, outputWriter{stream: "stderr", sink: sink})
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// OutputSink receives incremental process output chunks.
type OutputSink func(stream string, chunk string)

// CommandRunner abstracts process execution for tests.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) (models.ExecutionResult, error)
	RunShell(ctx context.Context, command string) (models.ExecutionResult, error)
}

// StreamingCommandRunner emits output while the process is still running.
type StreamingCommandRunner interface {
	RunStreaming(ctx context.Context, name string, sink OutputSink, args ...string) (models.ExecutionResult, error)
	RunShellStreaming(ctx context.Context, command string, sink OutputSink) (models.ExecutionResult, error)
}

// OSRunner executes commands on the current machine.
type OSRunner struct{}

// Run executes a binary with arguments.
func (OSRunner) Run(ctx context.Context, name string, args ...string) (models.ExecutionResult, error) {
	stdout, stderr, err := runProcess(ctx, name, args...)
	command := name
	if len(args) > 0 {
		command += " " + joinArgs(args)
	}

	result := models.ExecutionResult{
		Command:  command,
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode(err),
	}
	if err != nil {
		return result, fmt.Errorf("run command: %w", err)
	}

	return result, nil
}

// RunShell executes a shell command string.
func (runner OSRunner) RunShell(ctx context.Context, command string) (models.ExecutionResult, error) {
	result, err := runner.Run(ctx, "bash", "-c", command)
	result.Command = command
	if err != nil {
		return result, fmt.Errorf("run shell command: %w", err)
	}

	return result, nil
}

// RunStreaming executes a binary and reports output chunks as they arrive.
func (OSRunner) RunStreaming(ctx context.Context, name string, sink OutputSink, args ...string) (models.ExecutionResult, error) {
	stdout, stderr, err := runProcessStreaming(ctx, name, sink, args...)
	command := name
	if len(args) > 0 {
		command += " " + joinArgs(args)
	}

	result := models.ExecutionResult{
		Command:  command,
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode(err),
	}
	if err != nil {
		return result, fmt.Errorf("run command: %w", err)
	}

	return result, nil
}

// RunShellStreaming executes a shell command and reports output chunks as they arrive.
func (runner OSRunner) RunShellStreaming(ctx context.Context, command string, sink OutputSink) (models.ExecutionResult, error) {
	result, err := runner.RunStreaming(ctx, "bash", sink, "-c", command)
	result.Command = command
	if err != nil {
		return result, fmt.Errorf("run shell command: %w", err)
	}

	return result, nil
}

// joinArgs produces a display string for executed arguments.
func joinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}

	result := args[0]
	for i := 1; i < len(args); i++ {
		result += " " + args[i]
	}

	return result
}

// exitCode resolves a process exit code from an error.
func exitCode(err error) int {
	if err == nil {
		return 0
	}

	type exitCoder interface{ ExitCode() int }
	var exitErr exitCoder
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}

	return -1
}

type outputWriter struct {
	stream string
	sink   OutputSink
}

func (w outputWriter) Write(chunk []byte) (int, error) {
	if w.sink != nil && len(chunk) > 0 {
		w.sink(w.stream, string(append([]byte(nil), chunk...)))
	}

	return len(chunk), nil
}
