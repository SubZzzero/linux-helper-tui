package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"linux-helper/internal/app"
)

// main bootstraps the application and starts the TUI.
func main() {
	model, closeLog, err := app.Bootstrap()
	if err != nil {
		writeStderr("bootstrap failed: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if closeLog != nil {
			if err := closeLog(); err != nil {
				writeStderr("close log failed: %v\n", err)
			}
		}
	}()

	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		writeStderr("program failed: %v\n", err)
		os.Exit(1)
	}
}

// writeStderr prints a best-effort message to standard error.
func writeStderr(format string, args ...interface{}) {
	if _, err := fmt.Fprintf(os.Stderr, format, args...); err != nil {
		return
	}
}
