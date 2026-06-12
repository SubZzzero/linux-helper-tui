package app

import (
	"sort"

	tea "github.com/charmbracelet/bubbletea"

	"linux-helper/internal/i18n"
	"linux-helper/internal/models"
	"linux-helper/internal/storage"
	"linux-helper/internal/tui/screens"
	uitheme "linux-helper/internal/tui/theme"
)

type preferenceSaveMsg struct {
	config storage.Config
	err    error
}

// ConfigurePreferences wires the runtime locale and theme switching resources.
func (m *Model) ConfigurePreferences(translations map[string]models.LocalizedText, localeOrder []string, themes map[string]uitheme.Definition, currentTheme string, saveConfig func(storage.Config) error) {
	m.translations = translations
	m.localeOrder = append([]string(nil), localeOrder...)
	m.themes = make(map[string]uitheme.Definition, len(themes))
	for name, definition := range themes {
		m.themes[name] = definition
	}
	m.themeOrder = uitheme.Names(themes)
	m.currentTheme = currentTheme
	m.saveConfig = saveConfig
}

func (m *Model) cycleLocale() tea.Cmd {
	if len(m.localeOrder) < 2 || m.preferenceSave {
		return nil
	}

	return m.savePreferences(storage.Config{
		Locale: nextValue(m.localeOrder, m.locale),
		Theme:  m.currentTheme,
	})
}

func (m *Model) cycleTheme() tea.Cmd {
	if len(m.themeOrder) < 2 || m.preferenceSave {
		return nil
	}

	return m.savePreferences(storage.Config{
		Locale: m.locale,
		Theme:  nextValue(m.themeOrder, m.currentTheme),
	})
}

func (m *Model) savePreferences(config storage.Config) tea.Cmd {
	m.preferenceSave = true
	save := m.saveConfig
	return func() tea.Msg {
		if save != nil {
			return preferenceSaveMsg{config: config, err: save(config)}
		}

		return preferenceSaveMsg{config: config}
	}
}

func (m *Model) applyPreferences(config storage.Config) {
	m.locale = config.Locale
	if definition, ok := m.themes[config.Theme]; ok {
		m.styles = uitheme.NewStyles(definition)
		m.currentTheme = config.Theme
	}

	m.refreshRootScreen()
	m.refreshTopScreen()
}

func (m *Model) refreshRootScreen() {
	root := m.refreshScreen(m.stack.Root())
	m.stack.ReplaceRoot(m.sizeScreen(root))
}

func (m *Model) refreshTopScreen() {
	top := m.refreshScreen(m.stack.Top())
	m.stack.ReplaceTop(m.sizeScreen(top))
}

func (m *Model) refreshScreen(screen tea.Model) tea.Model {
	if len(m.translations) == 0 && len(m.themes) == 0 {
		return screen
	}

	translator := i18n.NewTranslator(m.locale, m.translations)
	switch current := screen.(type) {
	case screens.CatalogModel:
		current.SetPresentation(
			m.locale,
			m.styles,
			translator.T("app.title"),
			translator.T("catalog.empty"),
			translator.T("catalog.recent_title"),
			translator.T("catalog.recent_empty"),
			translator.T("catalog.help"),
		)
		return current
	case screens.DetailModel:
		current.SetPresentation(
			m.locale,
			m.styles,
			translator.T("detail.run"),
			translator.T("detail.back"),
			translator.T("detail.favorite"),
			translator.T("detail.favorite_on"),
			translator.T("detail.favorite_off"),
		)
		return current
	case screens.FormModel:
		current.SetPresentation(
			m.locale,
			m.styles,
			translator.T("form.preview"),
			translator.T("form.submit"),
			translator.T("form.back"),
		)
		return current
	case screens.ConfirmModel:
		current.SetPresentation(
			m.locale,
			m.styles,
			translator.T("confirm.title"),
			translator.T("confirm.approve"),
			translator.T("confirm.back"),
		)
		return current
	case screens.ResultModel:
		current.SetPresentation(
			m.locale,
			m.styles,
			translator.T("result.running"),
			translator.T("result.done"),
			translator.T("result.back"),
			translator.T("result.scroll"),
		)
		return current
	default:
		return screen
	}
}

func localeNames(translations map[string]models.LocalizedText) []string {
	namesSet := make(map[string]struct{})
	for _, text := range translations {
		for locale := range text {
			namesSet[locale] = struct{}{}
		}
	}

	names := make([]string, 0, len(namesSet))
	for locale := range namesSet {
		names = append(names, locale)
	}

	sort.Strings(names)
	return names
}

func nextValue(values []string, current string) string {
	if len(values) == 0 {
		return current
	}

	for index, value := range values {
		if value == current {
			return values[(index+1)%len(values)]
		}
	}

	return values[0]
}
