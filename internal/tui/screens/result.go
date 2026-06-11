package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"linux-helper/internal/models"
	uitheme "linux-helper/internal/tui/theme"
)

// ResultModel shows execution progress and the final process output.
type ResultModel struct {
	recipe      models.Recipe
	locale      string
	styles      uitheme.Styles
	runningText string
	doneText    string
	backText    string
	result      models.ExecutionResult
	err         error
	running     bool
	pendingBack bool
	width       int
}

// NewResultModel constructs an execution result screen in running state.
func NewResultModel(recipe models.Recipe, locale string, styles uitheme.Styles, runningText string, doneText string, backText string) ResultModel {
	return ResultModel{
		recipe:      recipe,
		locale:      locale,
		styles:      styles,
		runningText: runningText,
		doneText:    doneText,
		backText:    backText,
		running:     true,
	}
}

// Init starts the result screen with no async work.
func (m ResultModel) Init() tea.Cmd {
	return nil
}

// Update handles window size changes and exit from the result view.
func (m ResultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
	case tea.KeyMsg:
		switch typed.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter", "esc", "backspace", "q":
			if !m.running {
				m.pendingBack = true
			}
		}
	}

	return m, nil
}

// ConsumeBack reports whether the result screen requested a pop.
func (m *ResultModel) ConsumeBack() bool {
	back := m.pendingBack
	m.pendingBack = false
	return back
}

// SetOutcome stores the finished execution result.
func (m *ResultModel) SetOutcome(result models.ExecutionResult, err error) {
	m.result = result
	m.err = err
	m.running = false
}

// View renders the current execution state.
func (m ResultModel) View() string {
	lines := []string{
		m.styles.Title.Render(resolveRecipeText(m.locale, m.recipe.Title)),
		"",
	}

	if m.running {
		lines = append(lines, m.styles.Accent.Render(m.runningText))
		return renderFrame(m.styles, m.width, lines)
	}

	statusStyle := m.styles.Success
	if m.err != nil {
		statusStyle = m.styles.Error
	}

	lines = append(lines,
		statusStyle.Render(m.doneText),
		fmt.Sprintf("Command: %s", m.result.Command),
		fmt.Sprintf("Exit code: %d", m.result.ExitCode),
	)

	if m.result.Stdout != "" {
		lines = append(lines, "", m.styles.Accent.Render("Stdout:"), m.result.Stdout)
	}

	if m.result.Stderr != "" {
		lines = append(lines, "", m.styles.Accent.Render("Stderr:"), m.result.Stderr)
	}

	if m.err != nil {
		lines = append(lines, "", m.styles.Error.Render("Error: "+m.err.Error()))
	}

	lines = append(lines, "", m.styles.Muted.Render(m.backText))
	return renderFrame(m.styles, m.width, lines)
}
