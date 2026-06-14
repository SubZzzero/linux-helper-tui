package app

import (
	"context"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"linux-helper/internal/models"
	"linux-helper/internal/storage"
	"linux-helper/internal/tui/navigation"
	"linux-helper/internal/tui/screens"
	uitheme "linux-helper/internal/tui/theme"
)

// Executor runs one recipe with resolved field values.
type Executor interface {
	Execute(ctx context.Context, recipe models.Recipe, values map[string]string, confirmed bool) (models.ExecutionResult, error)
}

// StreamingExecutor emits output while a recipe is still running.
type StreamingExecutor interface {
	ExecuteStreaming(ctx context.Context, recipe models.Recipe, values map[string]string, confirmed bool, sink func(stream string, chunk string)) (models.ExecutionResult, error)
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
	id     int
	result models.ExecutionResult
	err    error
}

type executionOutputMsg struct {
	id     int
	stream string
	chunk  string
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
	resultCancel      string
	resultDone        string
	resultBack        string
	resultScroll      string
	executor          Executor
	pendingExecution  *executionRequest
	activeExecutionID int
	nextExecutionID   int
	cancelExecution   context.CancelFunc
	executionUpdates  <-chan tea.Msg
	windowWidth       int
	windowHeight      int
	logger            *slog.Logger
	translations      map[string]models.LocalizedText
	localeOrder       []string
	themes            map[string]uitheme.Definition
	themeOrder        []string
	currentTheme      string
	saveConfig        func(storage.Config) error
	preferenceSave    bool
}

// NewModel constructs the root model with one catalog screen.
func NewModel(catalog screens.CatalogModel, locale string, styles uitheme.Styles, favorites Favorites, recent RecentLoader, favoriteIDs []string, executor Executor, log *slog.Logger) Model {
	return Model{
		stack:             navigation.NewStack(catalog),
		locale:            locale,
		styles:            styles,
		favorites:         favorites,
		recent:            recent,
		favoriteIDs:       favoriteSet(favoriteIDs),
		detailRun:         "Press enter to fill fields and run",
		detailBack:        "Press esc to go back",
		detailFavorite:    "Favorite",
		detailFavoriteOn:  "Press ctrl+f to remove from favorites",
		detailFavoriteOff: "Press ctrl+f to add to favorites",
		formPreview:       "Preview",
		formSubmit:        "Press enter to continue",
		formBack:          "Press esc to go back",
		confirmTitle:      "Confirmation required",
		confirmApprove:    "Press enter to continue",
		confirmBack:       "Press esc to cancel",
		resultRunning:     "Running command...",
		resultCancel:      "Press esc to stop and go back",
		resultDone:        "Execution finished",
		resultScroll:      "Use up/down or pgup/pgdn to scroll",
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
	if sized, ok := msg.(tea.WindowSizeMsg); ok {
		m.windowWidth = sized.Width
		m.windowHeight = sized.Height
	}

	if finished, ok := msg.(executionFinishedMsg); ok {
		m.finishExecution(finished)
		return m, nil
	}

	if output, ok := msg.(executionOutputMsg); ok {
		if output.id != 0 && output.id == m.activeExecutionID {
			if resultScreen, ok := m.stack.Top().(screens.ResultModel); ok {
				resultScreen.AppendOutput(output.stream, output.chunk)
				m.stack.ReplaceTop(resultScreen)
			}
		}

		return m, waitForExecutionUpdate(m.executionUpdates)
	}

	if saved, ok := msg.(preferenceSaveMsg); ok {
		m.preferenceSave = false
		if saved.err != nil {
			if m.logger != nil {
				m.logger.Error("save preferences", "error", saved.err)
			}
			return m, nil
		}

		m.applyPreferences(saved.config)
		return m, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+l":
			if cmd := m.cycleLocale(); cmd != nil {
				return m, cmd
			}
		case "ctrl+t":
			if cmd := m.cycleTheme(); cmd != nil {
				return m, cmd
			}
		}
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
	case screens.CatalogModel:
		catalogScreen := current
		if category, ok := catalogScreen.ConsumeCategorySelection(); ok {
			catalogScreen.SetSelectedCategory(category)
			m.stack.ReplaceTop(catalogScreen)
			return nil
		}
		if recipe, ok := catalogScreen.ConsumeToggleFavorite(); ok {
			m.stack.ReplaceTop(catalogScreen)
			m.toggleFavoriteRecipe(recipe.ID)
			return nil
		}
		if recipe, ok := catalogScreen.ConsumeSelection(); ok {
			m.stack.ReplaceTop(catalogScreen)
			m.stack.Push(m.sizeScreen(screens.NewDetailModel(recipe, m.locale, m.styles, m.isFavorite(recipe.ID), m.detailRun, m.detailBack, m.detailFavorite, m.detailFavoriteOn, m.detailFavoriteOff)))
			if m.logger != nil {
				m.logger.Info("open recipe detail", "recipe_id", recipe.ID)
			}
			return nil
		}
		m.stack.ReplaceTop(catalogScreen)
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
			m.stack.Push(m.sizeScreen(screens.NewFormModel(recipe, m.locale, m.styles, m.formPreview, m.formSubmit, m.formBack)))
			return nil
		}
		if detailScreen.ConsumeBack() {
			m.stack.Pop()
			m.refreshTopScreen()
			return nil
		}
		m.stack.ReplaceTop(detailScreen)
	case screens.FormModel:
		formScreen := current
		if formScreen.ConsumeBack() {
			m.stack.Pop()
			m.refreshTopScreen()
			return nil
		}
		if values, ok := formScreen.ConsumeSubmit(); ok {
			recipe := formScreen.Recipe()
			request := &executionRequest{recipe: recipe, values: values, confirmed: recipe.Risk != models.RiskDangerous}
			m.pendingExecution = request
			m.stack.ReplaceTop(formScreen)
			if recipe.Risk == models.RiskDangerous {
				m.stack.Push(m.sizeScreen(screens.NewConfirmModel(recipe, m.locale, m.styles, formScreen.Preview(), m.confirmTitle, m.confirmApprove, m.confirmBack)))
				return nil
			}

			m.stack.Push(m.sizeScreen(screens.NewResultModel(recipe, m.locale, m.styles, m.resultRunning, m.resultCancel, m.resultDone, m.resultBack, m.resultScroll)))
			return m.executePending()
		}
		m.stack.ReplaceTop(formScreen)
	case screens.ConfirmModel:
		confirmScreen := current
		if confirmScreen.ConsumeBack() {
			m.pendingExecution = nil
			m.stack.Pop()
			m.refreshTopScreen()
			return nil
		}
		if confirmScreen.ConsumeConfirm() {
			if m.pendingExecution == nil {
				m.stack.Pop()
				return nil
			}
			m.stack.ReplaceTop(confirmScreen)
			m.stack.Push(m.sizeScreen(screens.NewResultModel(m.pendingExecution.recipe, m.locale, m.styles, m.resultRunning, m.resultCancel, m.resultDone, m.resultBack, m.resultScroll)))
			return m.executePending()
		}
		m.stack.ReplaceTop(confirmScreen)
	case screens.ResultModel:
		resultScreen := current
		if resultScreen.Running() && resultScreen.ConsumeStop() {
			m.cancelActiveExecution()
			m.stack.Pop()
			m.refreshTopScreen()
			return nil
		}
		if resultScreen.ConsumeBack() {
			m.stack.Pop()
			m.refreshTopScreen()
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
	m.nextExecutionID++
	executionID := m.nextExecutionID
	m.activeExecutionID = executionID
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelExecution = cancel
	m.executionUpdates = nil

	if streamer, ok := m.executor.(StreamingExecutor); ok {
		updates := make(chan tea.Msg, 128)
		m.executionUpdates = updates
		go func() {
			result, err := streamer.ExecuteStreaming(ctx, request.recipe, request.values, request.confirmed, func(stream string, chunk string) {
				select {
				case updates <- executionOutputMsg{id: executionID, stream: stream, chunk: chunk}:
				case <-ctx.Done():
				}
			})
			updates <- executionFinishedMsg{id: executionID, result: result, err: err}
			close(updates)
		}()

		return waitForExecutionUpdate(updates)
	}

	return func() tea.Msg {
		result, err := m.executor.Execute(ctx, request.recipe, request.values, request.confirmed)
		return executionFinishedMsg{id: executionID, result: result, err: err}
	}
}

func (m *Model) finishExecution(msg executionFinishedMsg) {
	if msg.id == 0 || msg.id != m.activeExecutionID {
		return
	}

	m.activeExecutionID = 0
	m.cancelExecution = nil
	m.executionUpdates = nil
	resultScreen, ok := m.stack.Top().(screens.ResultModel)
	if !ok {
		return
	}

	resultScreen.SetOutcome(msg.result, msg.err)
	m.stack.ReplaceTop(resultScreen)
	m.syncRecentCommands()
}

func (m *Model) cancelActiveExecution() {
	if m.cancelExecution != nil {
		m.cancelExecution()
	}

	m.cancelExecution = nil
	m.activeExecutionID = 0
	m.executionUpdates = nil
}

func waitForExecutionUpdate(updates <-chan tea.Msg) tea.Cmd {
	if updates == nil {
		return nil
	}

	return func() tea.Msg {
		msg, ok := <-updates
		if !ok {
			return nil
		}

		return msg
	}
}

// sizeScreen applies the latest known terminal geometry to a newly opened screen.
func (m Model) sizeScreen(screen tea.Model) tea.Model {
	if m.windowWidth == 0 || m.windowHeight == 0 {
		return screen
	}

	updated, _ := screen.Update(tea.WindowSizeMsg{Width: m.windowWidth, Height: m.windowHeight})
	return updated
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

	m.syncCatalogFavorites()
	return isFavorite, true
}

func (m *Model) syncCatalogFavorites() {
	catalogScreen, ok := m.stack.Root().(screens.CatalogModel)
	if !ok {
		return
	}

	catalogScreen.SetFavorites(favoriteIDs(m.favoriteIDs))
	m.stack.ReplaceRoot(catalogScreen)
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

	catalogScreen, ok := m.stack.Root().(screens.CatalogModel)
	if !ok {
		return
	}

	catalogScreen.SetRecent(recent)
	m.stack.ReplaceRoot(catalogScreen)
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
