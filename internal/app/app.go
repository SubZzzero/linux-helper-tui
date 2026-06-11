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

// Favorites toggles and loads favorite recipe identifiers.
type Favorites interface {
	Load() ([]string, error)
	Toggle(recipeID string) (bool, error)
}

// RecentLoader loads recently executed commands from persistence.
type RecentLoader interface {
	Load() ([]string, error)
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
	stack             *navigation.Stack
	locale            string
	styles            uitheme.Styles
	favorites         Favorites
	recent            RecentLoader
	favoriteIDs       map[string]struct{}
	detailRun         string
	detailBack        string
	detailFavorite    string
	detailFavoriteOn  string
	detailFavoriteOff string
	formPreview       string
	formSubmit        string
	formBack          string
	confirmTitle      string
	confirmApprove    string
	confirmBack       string
	resultRunning     string
	resultDone        string
	resultBack        string
	executor          Executor
	pendingExecution  *executionRequest
	logger            *slog.Logger
}

// NewModel constructs the root model with one search screen.
func NewModel(search screens.SearchModel, locale string, styles uitheme.Styles, favorites Favorites, recent RecentLoader, favoriteIDs []string, executor Executor, log *slog.Logger) Model {
	return Model{
		stack:             navigation.NewStack(search),
		locale:            locale,
		styles:            styles,
		favorites:         favorites,
		recent:            recent,
		favoriteIDs:       favoriteSet(favoriteIDs),
		detailRun:         "Press enter to fill fields and run",
		detailBack:        "Press esc to go back",
		detailFavorite:    "Favorite",
		detailFavoriteOn:  "Press f to remove from favorites",
		detailFavoriteOff: "Press f to add to favorites",
		formPreview:       "Preview",
		formSubmit:        "Press enter to continue",
		formBack:          "Press esc to go back",
		confirmTitle:      "Confirmation required",
		confirmApprove:    "Press enter or y to continue",
		confirmBack:       "Press esc, backspace, or n to cancel",
		resultRunning:     "Running command...",
		resultDone:        "Execution finished",
		resultBack:        "Press enter or esc to return",
		executor:          executor,
		logger:            log,
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
		if recipe, ok := searchScreen.ConsumeToggleFavorite(); ok {
			m.stack.ReplaceTop(searchScreen)
			m.toggleFavoriteRecipe(recipe.ID)
			return nil
		}
		if recipe, ok := searchScreen.ConsumeSelection(); ok {
			m.stack.ReplaceTop(searchScreen)
			m.stack.Push(screens.NewDetailModel(recipe, m.locale, m.styles, m.isFavorite(recipe.ID), m.detailRun, m.detailBack, m.detailFavorite, m.detailFavoriteOn, m.detailFavoriteOff))
			if m.logger != nil {
				m.logger.Info("open recipe detail", "recipe_id", recipe.ID)
			}
			return nil
		}
		m.stack.ReplaceTop(searchScreen)
	case screens.DetailModel:
		detailScreen := current
		if detailScreen.ConsumeToggleFavorite() {
			m.toggleFavorite(&detailScreen)
			m.stack.ReplaceTop(detailScreen)
			return nil
		}
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
	m.syncRecentCommands()
}

func (m *Model) isFavorite(recipeID string) bool {
	_, ok := m.favoriteIDs[recipeID]
	return ok
}

func (m *Model) toggleFavorite(detailScreen *screens.DetailModel) {
	recipeID := detailScreen.Recipe().ID
	if isFavorite, ok := m.toggleFavoriteRecipe(recipeID); ok {
		detailScreen.SetFavorite(isFavorite)
	}
}

func (m *Model) toggleFavoriteRecipe(recipeID string) (bool, bool) {
	if m.favorites == nil {
		return false, false
	}

	isFavorite, err := m.favorites.Toggle(recipeID)
	if err != nil {
		if m.logger != nil {
			m.logger.Error("toggle favorite", "recipe_id", recipeID, "error", err)
		}

		return false, false
	}

	if isFavorite {
		m.favoriteIDs[recipeID] = struct{}{}
	} else {
		delete(m.favoriteIDs, recipeID)
	}

	m.syncSearchFavorites()
	return isFavorite, true
}

func (m *Model) syncSearchFavorites() {
	searchScreen, ok := m.stack.Root().(screens.SearchModel)
	if !ok {
		return
	}

	searchScreen.SetFavorites(favoriteIDs(m.favoriteIDs))
	m.stack.ReplaceRoot(searchScreen)
}

func (m *Model) syncRecentCommands() {
	if m.recent == nil {
		return
	}

	recent, err := m.recent.Load()
	if err != nil {
		if m.logger != nil {
			m.logger.Error("load recent commands", "error", err)
		}
		return
	}

	searchScreen, ok := m.stack.Root().(screens.SearchModel)
	if !ok {
		return
	}

	searchScreen.SetRecent(recent)
	m.stack.ReplaceRoot(searchScreen)
}

func favoriteSet(favorites []string) map[string]struct{} {
	set := make(map[string]struct{}, len(favorites))
	for _, recipeID := range favorites {
		set[recipeID] = struct{}{}
	}

	return set
}

func favoriteIDs(favorites map[string]struct{}) []string {
	ids := make([]string, 0, len(favorites))
	for recipeID := range favorites {
		ids = append(ids, recipeID)
	}

	return ids
}
