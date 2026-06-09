package app

import (
	"context"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"linux-helper/internal/models"
	"linux-helper/internal/tui/navigation"
	"linux-helper/internal/tui/screens"
	uitheme "linux-helper/internal/tui/theme"
)

// Executor runs one recipe with resolved field values.
type Executor interface {
	Execute(ctx context.Context, recipe models.Recipe, values map[string]string, confirmed bool) (models.ExecutionResult, error)
}

type executionRequest struct {
	recipe    models.Recipe
	values    map[string]string
	confirmed bool
}

type executionFinishedMsg struct {
	result models.ExecutionResult
	err    error
}

// Model is the root Bubble Tea application model.
type Model struct {
	stack            *navigation.Stack
	locale           string
	styles           uitheme.Styles
	detailRun        string
	detailBack       string
	formPreview      string
	formSubmit       string
	formBack         string
	confirmTitle     string
	confirmApprove   string
	confirmBack      string
	resultRunning    string
	resultDone       string
	resultBack       string
	executor         Executor
	pendingExecution *executionRequest
	logger           *slog.Logger
}

// NewModel constructs the root model with one search screen.
func NewModel(search screens.SearchModel, locale string, styles uitheme.Styles, executor Executor, log *slog.Logger) Model {
	return Model{
		stack:          navigation.NewStack(search),
		locale:         locale,
		styles:         styles,
		detailRun:      "Press enter to fill fields and run",
		detailBack:     "Press esc to go back",
		formPreview:    "Preview",
		formSubmit:     "Press enter to continue",
		formBack:       "Press esc to go back",
		confirmTitle:   "Confirmation required",
		confirmApprove: "Press enter or y to continue",
		confirmBack:    "Press esc, backspace, or n to cancel",
		resultRunning:  "Running command...",
		resultDone:     "Execution finished",
		resultBack:     "Press enter or esc to return",
		executor:       executor,
		logger:         log,
	}
}

// Init delegates initialization to the top screen.
func (m Model) Init() tea.Cmd {
	return m.stack.Top().Init()
}

// Update updates the active screen and handles navigation transitions.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if finished, ok := msg.(executionFinishedMsg); ok {
		m.finishExecution(finished)
		return m, nil
	}

	updated, cmd := m.stack.Top().Update(msg)
	m.stack.ReplaceTop(updated)
	transitionCmd := m.handleTransitions()
	return m, tea.Batch(cmd, transitionCmd)
}

// View renders the active screen.
func (m Model) View() string {
	return m.stack.Top().View()
}

// handleTransitions applies screen-level navigation requests.
func (m *Model) handleTransitions() tea.Cmd {
	switch current := m.stack.Top().(type) {
	case screens.SearchModel:
		searchScreen := current
		if recipe, ok := searchScreen.ConsumeSelection(); ok {
			m.stack.ReplaceTop(searchScreen)
			m.stack.Push(screens.NewDetailModel(recipe, m.locale, m.styles, m.detailRun, m.detailBack))
			if m.logger != nil {
				m.logger.Info("open recipe detail", "recipe_id", recipe.ID)
			}
			return nil
		}
		m.stack.ReplaceTop(searchScreen)
	case screens.DetailModel:
		detailScreen := current
		if detailScreen.ConsumeExecute() {
			recipe := detailScreen.Recipe()
			m.stack.ReplaceTop(detailScreen)
			m.stack.Push(screens.NewFormModel(recipe, m.locale, m.styles, m.formPreview, m.formSubmit, m.formBack))
			return nil
		}
		if detailScreen.ConsumeBack() {
			m.stack.Pop()
			return nil
		}
		m.stack.ReplaceTop(detailScreen)
	case screens.FormModel:
		formScreen := current
		if formScreen.ConsumeBack() {
			m.stack.Pop()
			return nil
		}
		if values, ok := formScreen.ConsumeSubmit(); ok {
			recipe := formScreen.Recipe()
			request := &executionRequest{recipe: recipe, values: values, confirmed: recipe.Risk != models.RiskDangerous}
			m.pendingExecution = request
			m.stack.ReplaceTop(formScreen)
			if recipe.Risk == models.RiskDangerous {
				m.stack.Push(screens.NewConfirmModel(recipe, m.locale, m.styles, formScreen.Preview(), m.confirmTitle, m.confirmApprove, m.confirmBack))
				return nil
			}

			m.stack.Push(screens.NewResultModel(recipe, m.locale, m.styles, m.resultRunning, m.resultDone, m.resultBack))
			return m.executePending()
		}
		m.stack.ReplaceTop(formScreen)
	case screens.ConfirmModel:
		confirmScreen := current
		if confirmScreen.ConsumeBack() {
			m.pendingExecution = nil
			m.stack.Pop()
			return nil
		}
		if confirmScreen.ConsumeConfirm() {
			if m.pendingExecution == nil {
				m.stack.Pop()
				return nil
			}
			m.stack.ReplaceTop(confirmScreen)
			m.stack.Push(screens.NewResultModel(m.pendingExecution.recipe, m.locale, m.styles, m.resultRunning, m.resultDone, m.resultBack))
			return m.executePending()
		}
		m.stack.ReplaceTop(confirmScreen)
	case screens.ResultModel:
		resultScreen := current
		if resultScreen.ConsumeBack() {
			m.stack.Pop()
			return nil
		}
		m.stack.ReplaceTop(resultScreen)
	}

	return nil
}

func (m *Model) executePending() tea.Cmd {
	if m.pendingExecution == nil || m.executor == nil {
		return nil
	}

	request := *m.pendingExecution
	request.confirmed = true
	m.pendingExecution = nil

	return func() tea.Msg {
		result, err := m.executor.Execute(context.Background(), request.recipe, request.values, request.confirmed)
		return executionFinishedMsg{result: result, err: err}
	}
}

func (m *Model) finishExecution(msg executionFinishedMsg) {
	resultScreen, ok := m.stack.Top().(screens.ResultModel)
	if !ok {
		return
	}

	resultScreen.SetOutcome(msg.result, msg.err)
	m.stack.ReplaceTop(resultScreen)
}
