package app_test

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"linux-helper/internal/app"
	"linux-helper/internal/models"
	"linux-helper/internal/storage"
	"linux-helper/internal/tui/screens"
	uitheme "linux-helper/internal/tui/theme"
)

func appTestRecipes() []models.Recipe {
	return []models.Recipe{{
		ID:          "find-file",
		Version:     1,
		Category:    models.CategoryFilesystem,
		Risk:        models.RiskSafe,
		Execution:   models.ExecutionTypeDirect,
		Binary:      "find",
		Args:        []string{"{{path}}"},
		Fields:      []models.Field{{Name: "path", Type: models.FieldTypeString, Required: true, Default: "."}},
		Title:       models.LocalizedText{"en": "Find file"},
		Description: models.LocalizedText{"en": "Find files"},
	}, {
		ID:          "delete-tree",
		Version:     1,
		Category:    models.CategoryFilesystem,
		Risk:        models.RiskDangerous,
		Execution:   models.ExecutionTypeDirect,
		Binary:      "rm",
		Args:        []string{"-rf", "{{path}}"},
		Fields:      []models.Field{{Name: "path", Type: models.FieldTypeString, Required: true, Default: "/tmp/cache"}},
		Title:       models.LocalizedText{"en": "Delete tree", "ua": "Видалити дерево"},
		Description: models.LocalizedText{"en": "Delete a directory tree", "ua": "Видалити дерево каталогів"},
	}}
}

type fakeExecutor struct {
	result models.ExecutionResult
	err    error
	called int
	run    func(ctx context.Context, recipe models.Recipe, values map[string]string, confirmed bool) (models.ExecutionResult, error)
}

type fakeFavorites struct {
	ids map[string]struct{}
}

type fakeRecent struct {
	commands []string
}

// Execute records one execution request.
func (e *fakeExecutor) Execute(ctx context.Context, recipe models.Recipe, values map[string]string, confirmed bool) (models.ExecutionResult, error) {
	e.called++
	if e.run != nil {
		return e.run(ctx, recipe, values, confirmed)
	}
	return e.result, e.err
}

// Load returns the current favorite identifiers.
func (f *fakeFavorites) Load() ([]string, error) {
	ids := make([]string, 0, len(f.ids))
	for recipeID := range f.ids {
		ids = append(ids, recipeID)
	}

	return ids, nil
}

// Toggle flips one recipe identifier.
func (f *fakeFavorites) Toggle(recipeID string) (bool, error) {
	if _, ok := f.ids[recipeID]; ok {
		delete(f.ids, recipeID)
		return false, nil
	}

	f.ids[recipeID] = struct{}{}
	return true, nil
}

// Load returns the recent commands.
func (f *fakeRecent) Load() ([]string, error) {
	return append([]string(nil), f.commands...), nil
}

func appTestStyles() uitheme.Styles {
	return uitheme.NewStyles(uitheme.Definition{Name: "test", BorderColor: "63", AccentColor: "213"})
}

func appTestThemes() map[string]uitheme.Definition {
	return map[string]uitheme.Definition{
		"dark":  {Name: "dark", BorderColor: "63", AccentColor: "213"},
		"light": {Name: "light", BorderColor: "15", AccentColor: "10"},
	}
}

func appTestTranslations() map[string]models.LocalizedText {
	return map[string]models.LocalizedText{
		"app.title":            {"en": "linux-helper", "ua": "linux-helper", "ru": "linux-helper"},
		"catalog.empty":        {"en": "Empty", "ua": "Порожньо", "ru": "Пусто"},
		"catalog.recent_title": {"en": "Recent commands", "ua": "Останні команди", "ru": "Последние команды"},
		"catalog.recent_empty": {"en": "No recent commands yet.", "ua": "Ще немає команд.", "ru": "Команд пока нет."},
		"catalog.help":         {"en": "Catalog EN", "ua": "Каталог UA", "ru": "Каталог RU"},
		"detail.run":           {"en": "Detail EN", "ua": "Деталі UA", "ru": "Детали RU"},
		"detail.back":          {"en": "Back EN", "ua": "Назад UA", "ru": "Назад RU"},
		"detail.favorite":      {"en": "Favorite", "ua": "Обране", "ru": "Избранное"},
		"detail.favorite_on":   {"en": "Favorite on", "ua": "В обраному", "ru": "В избранном"},
		"detail.favorite_off":  {"en": "Favorite off", "ua": "Не в обраному", "ru": "Не в избранном"},
		"form.preview":         {"en": "Preview", "ua": "Попередній перегляд", "ru": "Предпросмотр"},
		"form.submit":          {"en": "Submit EN", "ua": "Надіслати UA", "ru": "Отправить RU"},
		"form.back":            {"en": "Form back EN", "ua": "Форма назад UA", "ru": "Форма назад RU"},
		"confirm.title":        {"en": "Confirm EN", "ua": "Підтвердити UA", "ru": "Подтвердить RU"},
		"confirm.approve":      {"en": "Approve EN", "ua": "Схвалити UA", "ru": "Подтвердить RU"},
		"confirm.back":         {"en": "Cancel EN", "ua": "Скасувати UA", "ru": "Отменить RU"},
		"result.running":       {"en": "Running EN", "ua": "Виконується UA", "ru": "Выполняется RU"},
		"result.cancel":        {"en": "Cancel running EN", "ua": "Скасувати виконання UA", "ru": "Прервать выполнение RU"},
		"result.done":          {"en": "Done EN", "ua": "Готово UA", "ru": "Готово RU"},
		"result.scroll":        {"en": "Scroll EN", "ua": "Прокрутка UA", "ru": "Прокрутка RU"},
		"result.back":          {"en": "Back result EN", "ua": "Назад результат UA", "ru": "Назад результат RU"},
	}
}

func appTestCatalog(recipes []models.Recipe, recent []string) screens.CatalogModel {
	return screens.NewCatalogModel(recipes, "en", appTestStyles(), nil, recent, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "Catalog EN")
}

func configureAppPreferences(model *app.Model, saved *[]storage.Config) {
	model.ConfigurePreferences(
		appTestTranslations(),
		[]string{"en", "ua", "ru"},
		appTestThemes(),
		"dark",
		func(config storage.Config) error {
			*saved = append(*saved, config)
			return nil
		},
	)
}

// TestModelView renders the active screen.
func TestModelView(t *testing.T) {
	catalogModel := screens.NewCatalogModel(appTestRecipes(), "en", appTestStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")
	model := app.NewModel(catalogModel, "en", appTestStyles(), nil, nil, nil, nil, nil)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotEmpty(t, updated.View())
}

// TestModelExecutesSafeRecipe drives the minimal execution flow.
func TestModelExecutesSafeRecipe(t *testing.T) {
	executor := &fakeExecutor{result: models.ExecutionResult{Command: "find .", ExitCode: 0, Stdout: "ok"}}
	recent := &fakeRecent{commands: []string{"find ."}}
	catalogModel := screens.NewCatalogModel(appTestRecipes(), "en", appTestStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")
	model := app.NewModel(catalogModel, "en", appTestStyles(), nil, recent, nil, executor, nil)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, updated)
	require.Nil(t, cmd)

	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, updated)
	require.Nil(t, cmd)

	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, updated)
	require.Nil(t, cmd)

	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, updated)
	require.NotNil(t, cmd)
	assert.Equal(t, 0, executor.called)

	updated, _ = updated.Update(cmd())
	assert.Equal(t, 1, executor.called)
	assert.Contains(t, updated.View(), "Execution finished")
	assert.Contains(t, updated.View(), "find .")

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Contains(t, updated.View(), "Recent commands")
	assert.Contains(t, updated.View(), "- find .")
}

// TestModelTogglesFavorites updates detail and catalog state.
func TestModelTogglesFavorites(t *testing.T) {
	favorites := &fakeFavorites{ids: map[string]struct{}{}}
	catalogModel := screens.NewCatalogModel(appTestRecipes(), "en", appTestStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")
	model := app.NewModel(catalogModel, "en", appTestStyles(), favorites, nil, nil, nil, nil)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyCtrlF})

	assert.Contains(t, updated.View(), "Favorite: yes")

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Contains(t, updated.View(), "[*] Find file")
}

// TestModelCarriesWindowSizeToResultScreen keeps result output usable without a manual resize.
func TestModelCarriesWindowSizeToResultScreen(t *testing.T) {
	executor := &fakeExecutor{result: models.ExecutionResult{Command: "ps aux", ExitCode: 0, Stdout: "header\nbody\nfooter"}}
	catalogModel := screens.NewCatalogModel(appTestRecipes(), "en", appTestStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")
	model := app.NewModel(catalogModel, "en", appTestStyles(), nil, nil, nil, executor, nil)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	updated, _ = updated.Update(cmd())
	view := updated.View()

	assert.Contains(t, view, "Execution finished")
	assert.Contains(t, view, "header")
	assert.Contains(t, view, "body")
	assert.Contains(t, view, "footer")
	assert.Contains(t, view, "Use up/down or pgup/pgdn to scroll")
}

// TestModelLocaleHotkeyCyclesCatalogWithoutResettingSelection keeps the category filter active.
func TestModelLocaleHotkeyCyclesCatalogWithoutResettingSelection(t *testing.T) {
	catalogModel := appTestCatalog(appTestRecipes(), []string{"find ."})
	model := app.NewModel(catalogModel, "en", appTestStyles(), nil, nil, nil, nil, nil)
	saved := []storage.Config{}
	configureAppPreferences(&model, &saved)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel := updated.(app.Model)
	updated, cmd := updatedModel.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	require.NotNil(t, cmd)
	updated, _ = updated.Update(cmd())
	view := updated.View()

	assert.Contains(t, view, "Останні команди")
	assert.Contains(t, view, "Каталог UA")
	assert.Contains(t, view, "Find file")
	assert.NotContains(t, view, "Disk usage")
	require.Len(t, saved, 1)
	assert.Equal(t, storage.Config{Locale: "ua", Theme: "dark"}, saved[0])

	updatedModel = updated.(app.Model)
	updated, cmd = updatedModel.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	require.NotNil(t, cmd)
	updated, _ = updated.Update(cmd())
	assert.Contains(t, updated.View(), "Последние команды")
}

// TestModelLocaleHotkeyUpdatesActiveFormWithoutLosingValues keeps typed form state intact.
func TestModelLocaleHotkeyUpdatesActiveFormWithoutLosingValues(t *testing.T) {
	recipes := []models.Recipe{{
		ID:          "find-file",
		Version:     1,
		Category:    models.CategoryFilesystem,
		Risk:        models.RiskSafe,
		Execution:   models.ExecutionTypeDirect,
		Binary:      "find",
		Args:        []string{"{{path}}"},
		Fields:      []models.Field{{Name: "path", Type: models.FieldTypeString, Required: true}},
		Title:       models.LocalizedText{"en": "Find file", "ua": "Знайти файл"},
		Description: models.LocalizedText{"en": "Find files", "ua": "Знайти файли"},
	}}
	catalogModel := appTestCatalog(recipes, nil)
	model := app.NewModel(catalogModel, "en", appTestStyles(), nil, nil, nil, nil, nil)
	saved := []storage.Config{}
	configureAppPreferences(&model, &saved)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Runes: []rune{'/'}, Type: tea.KeyRunes})

	updatedModel := updated.(app.Model)
	updated, cmd := updatedModel.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	require.NotNil(t, cmd)
	updated, _ = updated.Update(cmd())
	view := updated.View()

	assert.Contains(t, view, "Попередній перегляд")
	assert.Contains(t, view, "find /")
	assert.Contains(t, view, "Знайти файли")
	require.Len(t, saved, 1)
	assert.Equal(t, "ua", saved[0].Locale)
	assert.Equal(t, "dark", saved[0].Theme)
}

// TestModelCancelsRunningExecutionOnEscape returns to the previous screen without quitting.
func TestModelCancelsRunningExecutionOnEscape(t *testing.T) {
	canceled := make(chan struct{}, 1)
	executor := &fakeExecutor{
		run: func(ctx context.Context, _ models.Recipe, _ map[string]string, _ bool) (models.ExecutionResult, error) {
			<-ctx.Done()
			canceled <- struct{}{}
			return models.ExecutionResult{Command: "tail -f /var/log/syslog", ExitCode: -1}, ctx.Err()
		},
	}
	catalogModel := screens.NewCatalogModel(appTestRecipes(), "en", appTestStyles(), nil, nil, "linux-helper", "Empty", "Recent commands", "No recent commands yet.", "up/down move, enter open, esc back, ctrl+c quit")
	model := app.NewModel(catalogModel, "en", appTestStyles(), nil, nil, nil, executor, nil)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	go func() {
		_ = cmd()
	}()

	assert.Contains(t, updated.View(), "Running command")
	assert.Contains(t, updated.View(), "Press esc to stop and go back")
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Contains(t, updated.View(), "Preview")
	assert.Contains(t, updated.View(), "find .")
	assert.Equal(t, 1, executor.called)

	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("expected running command to be canceled")
	}
	assert.NotContains(t, updated.View(), "Execution finished")
	assert.NotContains(t, updated.View(), "context canceled")
	assert.NotContains(t, updated.View(), "tail -f /var/log/syslog")
}
