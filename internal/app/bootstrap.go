package app

import (
	"fmt"
	"io/fs"
	"os"

	linuxhelper "linux-helper"
	"linux-helper/internal/executor"
	"linux-helper/internal/i18n"
	"linux-helper/internal/logger"
	"linux-helper/internal/recipes"
	"linux-helper/internal/services"
	"linux-helper/internal/storage"
	"linux-helper/internal/tui/screens"
	uitheme "linux-helper/internal/tui/theme"
)

// Bootstrap constructs the root application model.
func Bootstrap() (Model, func() error, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Model{}, nil, fmt.Errorf("resolve home dir: %w", err)
	}

	paths := storage.DefaultPaths(home)
	config, err := storage.LoadConfig(paths.ConfigFile)
	if err != nil {
		return Model{}, nil, err
	}

	log, closer, err := logger.New(paths.LogFile)
	if err != nil {
		return Model{}, nil, err
	}

	localeFS, err := fs.Sub(linuxhelper.Assets, "assets/locales")
	if err != nil {
		return Model{}, nil, fmt.Errorf("open embedded locales: %w", err)
	}

	translations, err := i18n.LoadLocales(localeFS, ".")
	if err != nil {
		return Model{}, nil, err
	}

	translator := i18n.NewTranslator(config.Locale, translations)
	themeFS, err := fs.Sub(linuxhelper.Assets, "assets/themes")
	if err != nil {
		return Model{}, nil, fmt.Errorf("open embedded themes: %w", err)
	}

	themes, err := uitheme.LoadDefinitions(themeFS, ".")
	if err != nil {
		return Model{}, nil, err
	}

	definition, err := uitheme.ResolveDefinition(themes, config.Theme)
	if err != nil {
		return Model{}, nil, err
	}

	styles := uitheme.NewStyles(definition)

	recipeFS, err := fs.Sub(linuxhelper.Assets, "assets/recipes")
	if err != nil {
		return Model{}, nil, fmt.Errorf("open embedded recipes: %w", err)
	}

	var overrideFS fs.FS
	if info, statErr := os.Stat(paths.RecipesDir); statErr == nil && info.IsDir() {
		overrideFS = os.DirFS(paths.RecipesDir)
	}

	loader := recipes.NewLoader(recipeFS, overrideFS, ".")
	recipeService, err := services.NewRecipeService(loader)
	if err != nil {
		return Model{}, nil, err
	}

	searchService := services.NewSearchService(recipeService.All())
	searchScreen, err := screens.NewSearchModel(
		searchService,
		config.Locale,
		styles,
		translator.T("app.title"),
		translator.T("search.placeholder"),
		translator.T("search.empty"),
	)
	if err != nil {
		return Model{}, nil, err
	}

	recentStore := storage.NewRecentStore(paths.RecentFile)
	executionService := services.NewExecutionService(executor.OSRunner{}, recentStore)

	model := NewModel(searchScreen, config.Locale, styles, executionService, log)
	model.detailRun = translator.T("detail.run")
	model.detailBack = translator.T("detail.back")
	model.formPreview = translator.T("form.preview")
	model.formSubmit = translator.T("form.submit")
	model.formBack = translator.T("form.back")
	model.confirmTitle = translator.T("confirm.title")
	model.confirmApprove = translator.T("confirm.approve")
	model.confirmBack = translator.T("confirm.back")
	model.resultRunning = translator.T("result.running")
	model.resultDone = translator.T("result.done")
	model.resultBack = translator.T("result.back")
	return model, closer.Close, nil
}
