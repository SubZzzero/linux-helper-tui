package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"linux-helper/internal/models"
	uitheme "linux-helper/internal/tui/theme"
)

// ResultModel shows execution progress and the final process output.
type ResultModel struct {
	recipe      models.Recipe
	locale      string
	styles      uitheme.Styles
	runningText string
	cancelText  string
	doneText    string
	backText    string
	scrollText  string
	result      models.ExecutionResult
	err         error
	running     bool
	pendingBack bool
	pendingStop bool
	width       int
	height      int
	viewport    viewport.Model
}

// NewResultModel constructs an execution result screen in running state.
func NewResultModel(recipe models.Recipe, locale string, styles uitheme.Styles, runningText string, cancelText string, doneText string, backText string, scrollText string) ResultModel {
	return ResultModel{
		recipe:      recipe,
		locale:      locale,
		styles:      styles,
		runningText: runningText,
		cancelText:  cancelText,
		doneText:    doneText,
		backText:    backText,
		scrollText:  scrollText,
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
		m.height = typed.Height
		m.syncViewport()
	case tea.KeyMsg:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(typed)
		if !m.running {
			switch typed.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter", "esc":
				m.pendingBack = true
			}

			return m, cmd
		}

		switch typed.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.pendingStop = true
		}

		return m, cmd
	}

	return m, nil
}

// ConsumeStop reports whether the running command should be interrupted.
func (m *ResultModel) ConsumeStop() bool {
	stop := m.pendingStop
	m.pendingStop = false
	return stop
}

// Running reports whether the command is still executing.
func (m ResultModel) Running() bool {
	return m.running
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
	m.syncViewport()
}

// AppendOutput adds one live output chunk to the active result buffer.
func (m *ResultModel) AppendOutput(stream string, chunk string) {
	switch stream {
	case "stderr":
		m.result.Stderr += chunk
	default:
		m.result.Stdout += chunk
	}

	m.syncViewport()
}

// SetPresentation updates localized strings and styles without resetting state.
func (m *ResultModel) SetPresentation(locale string, styles uitheme.Styles, runningText string, cancelText string, doneText string, backText string, scrollText string) {
	m.locale = locale
	m.styles = styles
	m.runningText = runningText
	m.cancelText = cancelText
	m.doneText = doneText
	m.backText = backText
	m.scrollText = scrollText
	m.syncViewport()
}

// View renders the current execution state.
func (m ResultModel) View() string {
	title := m.styles.Title.Render(resolveRecipeText(m.locale, m.recipe.Title))

	if m.running {
		headLines := []string{title, "", m.styles.Accent.Render(m.runningText), m.styles.Muted.Render(m.cancelText)}
		bodyContent := m.viewport.View()
		if m.width == 0 || m.height == 0 {
			bodyContent = m.outputContent()
		}
		footLines := []string{"", m.styles.Muted.Render(m.scrollText)}
		body := strings.Join(append(append(headLines, bodyContent), footLines...), "\n")
		return resultFrame(m.styles, m.width).Render(body)
	}

	statusStyle := m.styles.Success
	if m.err != nil {
		statusStyle = m.styles.Error
	}

	headLines := []string{
		title,
		"",
		statusStyle.Render(m.doneText),
		fmt.Sprintf("Command: %s", m.result.Command),
		fmt.Sprintf("Exit code: %d", m.result.ExitCode),
	}

	footLines := []string{"", m.styles.Muted.Render(m.scrollText), m.styles.Muted.Render(m.backText)}
	body := strings.Join(append(append(headLines, m.viewport.View()), footLines...), "\n")
	return resultFrame(m.styles, m.width).Render(body)
}

// syncViewport keeps the output viewport aligned with the current result and window size.
func (m *ResultModel) syncViewport() {
	offset := m.viewport.YOffset
	contentWidth, contentHeight := m.viewportDimensions()
	vp := viewport.New(contentWidth, contentHeight)
	vp.SetContent(m.outputContent())
	if m.running {
		vp.GotoBottom()
	} else {
		vp.SetYOffset(offset)
	}
	m.viewport = vp
}

// viewportDimensions returns the inner viewport size after accounting for the frame and static text.
func (m ResultModel) viewportDimensions() (int, int) {
	frame := resultFrame(m.styles, m.width)
	horizontalFrame, verticalFrame := frame.GetFrameSize()
	frameWidth := 1
	if m.width > 0 {
		frameWidth = max(1, m.width-2)
	}

	contentWidth := max(1, frameWidth-horizontalFrame)
	staticLines := 8
	if m.running {
		staticLines = 6
	}
	availableHeight := max(1, m.height-verticalFrame-staticLines)
	return contentWidth, availableHeight
}

// outputContent joins the dynamic execution output into one scrollable text block.
func (m ResultModel) outputContent() string {
	sections := make([]string, 0, 3)
	if m.result.Stdout != "" {
		sections = append(sections, strings.Join([]string{m.styles.Accent.Render("Stdout:"), m.result.Stdout}, "\n"))
	}

	if m.result.Stderr != "" {
		sections = append(sections, strings.Join([]string{m.styles.Accent.Render("Stderr:"), m.result.Stderr}, "\n"))
	}

	if m.err != nil {
		sections = append(sections, m.styles.Error.Render("Error: "+m.err.Error()))
	}

	if len(sections) == 0 {
		if m.running {
			return m.styles.Muted.Render("No output yet")
		}

		return m.styles.Muted.Render("No output")
	}

	return strings.Join(sections, "\n\n")
}

// resultFrame returns the common frame style used by the result screen.
func resultFrame(styles uitheme.Styles, width int) lipgloss.Style {
	frame := styles.Frame
	if width > 0 {
		frame = frame.Width(max(1, width-2))
	}

	return frame
}
