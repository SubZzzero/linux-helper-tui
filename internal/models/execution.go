package models

import "fmt"

// ExecutionType controls how the command is launched.
type ExecutionType string

const (
	// ExecutionTypeDirect launches a binary with arguments.
	ExecutionTypeDirect ExecutionType = "direct"
	// ExecutionTypeShell launches a shell command string.
	ExecutionTypeShell ExecutionType = "shell"
)

// RiskLevel marks the command confirmation requirement.
type RiskLevel string

const (
	// RiskSafe runs without extra confirmation.
	RiskSafe RiskLevel = "safe"
	// RiskElevated indicates a moderate command risk.
	RiskElevated RiskLevel = "elevated"
	// RiskDangerous requires explicit confirmation.
	RiskDangerous RiskLevel = "dangerous"
)

// ExecutionResult captures the process outcome.
type ExecutionResult struct {
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
}

// Valid reports whether the execution type is supported.
func (e ExecutionType) Valid() bool {
	return e == ExecutionTypeDirect || e == ExecutionTypeShell
}

// Valid reports whether the risk level is supported.
func (r RiskLevel) Valid() bool {
	return r == RiskSafe || r == RiskElevated || r == RiskDangerous
}

// ParseExecutionType validates an execution type.
func ParseExecutionType(value string) (ExecutionType, error) {
	executionType := ExecutionType(value)
	if !executionType.Valid() {
		return "", fmt.Errorf("unknown execution type %q", value)
	}

	return executionType, nil
}

// ParseRiskLevel validates a risk level.
func ParseRiskLevel(value string) (RiskLevel, error) {
	riskLevel := RiskLevel(value)
	if !riskLevel.Valid() {
		return "", fmt.Errorf("unknown risk level %q", value)
	}

	return riskLevel, nil
}
