package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"linux-helper/internal/models"
	uitheme "linux-helper/internal/tui/theme"
)

// CatalogModel renders the recipe catalog screen.
type CatalogModel struct {
	locale           string
	styles           uitheme.Styles
	favorites        map[string]struct{}
	recent           []string
	allRecipes       []models.Recipe
	categories       []models.Category
	selectedCategory models.Category
	results          []models.Recipe
	selected         int
	pending          *models.Recipe
	pendingFavorite  *models.Recipe
	pendingCategory  *models.Category
	title            string
	emptyText        string
	recentTitle      string
	recentEmpty      string
	helpText         string
	width            int
	height           int
}

// NewCatalogModel constructs the initial recipe catalog screen.
func NewCatalogModel(recipes []models.Recipe, locale string, styles uitheme.Styles, favorites []string, recent []string, title string, emptyText string, recentTitle string, recentEmpty string, helpText string) CatalogModel {
	orderedRecipes := orderResults(recipes, favoriteSet(favorites))
	model := CatalogModel{
		locale:      locale,
		styles:      styles,
		favorites:   favoriteSet(favorites),
		recent:      append([]string(nil), recent...),
		allRecipes:  orderedRecipes,
		categories:  categoriesFromResults(orderedRecipes),
		title:       title,
		emptyText:   emptyText,
		recentTitle: recentTitle,
		recentEmpty: recentEmpty,
		helpText:    helpText,
	}
	model.applyCategoryFilter()
	return model
}

// Init starts the screen with no async work.
func (m CatalogModel) Init() tea.Cmd {
	return nil
}

// Update handles category and recipe selection changes.
func (m CatalogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
	case tea.KeyMsg:
		switch typed.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "backspace", "q":
			if m.selectedCategory != "" {
				m.selectedCategory = ""
				m.selected = 0
				m.applyCategoryFilter()
			}
			return m, nil
		case "up", "ctrl+p":
			if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "down", "ctrl+n":
			if m.selected < m.lastSelectableIndex() {
				m.selected++
			}
			return m, nil
		case "home":
			if m.lastSelectableIndex() >= 0 {
				m.selected = 0
			}
			return m, nil
		case "end":
			if m.lastSelectableIndex() >= 0 {
				m.selected = m.lastSelectableIndex()
			}
			return m, nil
		case "enter":
			if m.selectedCategory == "" {
				if len(m.categories) > 0 {
					selected := m.categories[m.selected]
					m.pendingCategory = &selected
				}
				return m, nil
			}

			if len(m.results) > 0 {
				selected := m.results[m.selected]
				m.pending = &selected
			}
			return m, nil
		}

		// Typed characters are intentionally ignored while the catalog is browse-only.
		return m, nil
	}

	return m, nil
}

// ConsumeToggleFavorite returns the selected recipe for a favorite toggle once.
func (m *CatalogModel) ConsumeToggleFavorite() (models.Recipe, bool) {
	if m.pendingFavorite == nil {
		return models.Recipe{}, false
	}

	recipe := *m.pendingFavorite
	m.pendingFavorite = nil
	return recipe, true
}

// ConsumeCategorySelection returns the selected category once.
func (m *CatalogModel) ConsumeCategorySelection() (models.Category, bool) {
	if m.pendingCategory == nil {
		return "", false
	}

	category := *m.pendingCategory
	m.pendingCategory = nil
	return category, true
}

// SetFavorites updates the favorite state used for rendering and ordering.
func (m *CatalogModel) SetFavorites(favorites []string) {
	m.favorites = favoriteSet(favorites)
	m.allRecipes = orderResults(m.allRecipes, m.favorites)
	m.applyCategoryFilter()
}

// SetRecent updates the recent command list.
func (m *CatalogModel) SetRecent(recent []string) {
	m.recent = append([]string(nil), recent...)
}

// SetSelectedCategory switches the active category filter.
func (m *CatalogModel) SetSelectedCategory(category models.Category) {
	m.selectedCategory = category
	m.selected = 0
	m.applyCategoryFilter()
}

// ConsumeSelection returns the selected recipe once.
func (m *CatalogModel) ConsumeSelection() (models.Recipe, bool) {
	if m.pending == nil {
		return models.Recipe{}, false
	}

	recipe := *m.pending
	m.pending = nil
	return recipe, true
}

// View renders the catalog UI.
func (m CatalogModel) View() string {
	content := m.layoutLines()
	return renderFrame(m.styles, m.width, content)
}

// renderResults renders the current result list.
func (m CatalogModel) renderResults() string {
	if m.selectedCategory == "" {
		return m.renderCategoryRows()
	}

	lines := make([]string, 0, len(m.results))
	contentWidth := m.availableContentWidth()
	for index, recipe := range m.results {
		line := truncateText("  "+favoriteMarker(m.favorites, recipe.ID)+" "+resolveRecipeText(m.locale, recipe.Title), contentWidth)
		if index == m.selected {
			line = m.styles.Selected.Render(truncateText("> "+favoriteMarker(m.favorites, recipe.ID)+" "+resolveRecipeText(m.locale, recipe.Title), contentWidth))
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m CatalogModel) renderCategoryRows() string {
	lines := make([]string, 0, len(m.categories))
	nameWidth := categoryNameWidth(m.categories)
	contentWidth := m.availableContentWidth()
	for index, category := range m.categories {
		line := truncateText("  "+categoryLine(m.locale, category, nameWidth), contentWidth)
		if index == m.selected {
			line = m.styles.Selected.Render(truncateText("> "+categoryLine(m.locale, category, nameWidth), contentWidth))
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m CatalogModel) layoutLines() []string {
	mainLines := []string{m.styles.Title.Render(m.title), ""}
	if m.isEmpty() {
		mainLines = append(mainLines, m.styles.Muted.Render(truncateText(m.emptyText, m.availableContentWidth())))
	} else {
		mainLines = append(mainLines, strings.Split(m.renderResults(), "\n")...)
	}

	availableHeight := m.availableContentHeight()
	if availableHeight <= 0 {
		return append(mainLines, m.fullFooterLines(maxRecentVisible)...)
	}

	helpLines := []string{"", m.styles.Muted.Render(m.helpText)}
	remainingHeight := availableHeight - len(mainLines)
	if remainingHeight < len(helpLines) {
		return trimLines(mainLines, availableHeight)
	}

	recentLimit := remainingHeight - len(helpLines) - 2
	if recentLimit < 0 {
		recentLimit = 0
	}

	content := append([]string(nil), mainLines...)
	content = append(content, m.footerLines(recentLimit)...)
	content = trimLines(content, availableHeight)
	return content
}

func (m CatalogModel) fullFooterLines(recentLimit int) []string {
	footer := m.footerLines(recentLimit)
	if len(footer) == 0 {
		return nil
	}

	return append([]string(nil), footer...)
}

func (m CatalogModel) footerLines(recentLimit int) []string {
	lines := []string{"", m.styles.Accent.Render(m.recentTitle)}
	if len(m.recent) == 0 || recentLimit == 0 {
		lines = append(lines, m.styles.Muted.Render(truncateText(m.recentEmpty, m.availableContentWidth())))
	} else {
		lines = append(lines, m.renderRecentLines(recentLimit)...)
	}

	lines = append(lines, "", m.styles.Muted.Render(truncateText(m.helpText, m.availableContentWidth())))
	return lines
}

func (m CatalogModel) renderRecentLines(limit int) []string {
	lines := make([]string, 0, min(limit, len(m.recent)))
	contentWidth := m.availableContentWidth()
	for _, command := range m.recent {
		if len(lines) >= limit {
			break
		}

		wrapped := wrapPrefixedText(command, contentWidth, "- ", "  ")
		remaining := limit - len(lines)
		if len(wrapped) > remaining {
			wrapped = wrapped[:remaining]
		}

		lines = append(lines, wrapped...)
	}

	return lines
}

func (m CatalogModel) availableContentHeight() int {
	if m.height <= 0 {
		return 0
	}

	return max(1, m.height-framedVerticalOverhead)
}

func (m CatalogModel) availableContentWidth() int {
	if m.width <= 0 {
		return 0
	}

	return max(1, m.width-framedHorizontalOverhead)
}

func trimLines(lines []string, limit int) []string {
	if limit <= 0 || len(lines) <= limit {
		return lines
	}

	trimmed := append([]string(nil), lines[:limit]...)
	if limit > 0 && len(lines) > limit {
		trimmed[limit-1] = strings.TrimRight(trimmed[limit-1], " ")
	}

	return trimmed
}

func favoriteSet(favorites []string) map[string]struct{} {
	set := make(map[string]struct{}, len(favorites))
	for _, favoriteID := range favorites {
		set[favoriteID] = struct{}{}
	}

	return set
}

func orderResults(results []models.Recipe, favorites map[string]struct{}) []models.Recipe {
	ordered := make([]models.Recipe, 0, len(results))
	for _, recipe := range results {
		if _, ok := favorites[recipe.ID]; ok {
			ordered = append(ordered, recipe)
		}
	}
	for _, recipe := range results {
		if _, ok := favorites[recipe.ID]; !ok {
			ordered = append(ordered, recipe)
		}
	}

	return ordered
}

func favoriteMarker(favorites map[string]struct{}, recipeID string) string {
	if _, ok := favorites[recipeID]; ok {
		return "[*]"
	}

	return "[ ]"
}

func (m *CatalogModel) applyCategoryFilter() {
	m.categories = categoriesFromResults(m.allRecipes)
	if m.selectedCategory != "" && !containsCategory(m.categories, m.selectedCategory) {
		m.selectedCategory = ""
	}

	filtered := make([]models.Recipe, 0, len(m.allRecipes))
	if m.selectedCategory == "" {
		m.results = filtered
		if m.selected >= len(m.categories) {
			m.selected = max(0, len(m.categories)-1)
		}
		return
	} else {
		for _, recipe := range m.allRecipes {
			if recipe.Category == m.selectedCategory {
				filtered = append(filtered, recipe)
			}
		}
	}

	m.results = filtered
	if m.selected >= len(m.results) {
		m.selected = max(0, len(m.results)-1)
	}
}

func (m CatalogModel) isEmpty() bool {
	if m.selectedCategory == "" {
		return len(m.categories) == 0
	}

	return len(m.results) == 0
}

func (m CatalogModel) lastSelectableIndex() int {
	if m.selectedCategory == "" {
		return len(m.categories) - 1
	}

	return len(m.results) - 1
}

func categoriesFromResults(results []models.Recipe) []models.Category {
	seen := make(map[models.Category]struct{}, len(results))
	categories := make([]models.Category, 0, len(results))
	for _, recipe := range results {
		if _, ok := seen[recipe.Category]; ok {
			continue
		}

		seen[recipe.Category] = struct{}{}
		categories = append(categories, recipe.Category)
	}

	return categories
}

func containsCategory(categories []models.Category, target models.Category) bool {
	for _, category := range categories {
		if category == target {
			return true
		}
	}

	return false
}

func categoryLine(locale string, category models.Category, nameWidth int) string {
	name := category.DisplayName()
	padding := strings.Repeat(" ", max(0, nameWidth-textWidth(name)))
	return name + padding + "   " + categoryDescription(locale, category)
}

func categoryNameWidth(categories []models.Category) int {
	width := 0
	for _, category := range categories {
		width = max(width, textWidth(category.DisplayName()))
	}

	return width
}

// categoryDescription returns one short localized hint for a category.
func categoryDescription(locale string, category models.Category) string {
	if descriptions, ok := localizedCategoryDescriptions[locale]; ok {
		if description, ok := descriptions[category]; ok {
			return description
		}
		return descriptions[""]
	}

	if description, ok := localizedCategoryDescriptions["en"][category]; ok {
		return description
	}

	return localizedCategoryDescriptions["en"][""]
}

var localizedCategoryDescriptions = map[string]map[models.Category]string{
	"en": {
		models.CategoryFilesystem:      "Files, directories, and permissions",
		models.CategoryEnvironment:     "Environment variables and shell",
		models.CategoryLogs:            "System logs and journal",
		models.CategoryNetwork:         "Network, ports, and connections",
		models.CategoryPackages:        "Packages and package managers",
		models.CategoryProcesses:       "Processes and system load",
		models.CategoryServices:        "Services and systemd",
		models.CategorySystem:          "System, disks, and resources",
		models.CategoryText:            "Search and text processing",
		models.CategoryTroubleshooting: "Failure triage, diagnostics, and root-cause checks",
		models.CategoryUsers:           "Users and sessions",
		"":                             "Category commands",
	},
	"ru": {
		models.CategoryFilesystem:      "Файлы, каталоги и права",
		models.CategoryEnvironment:     "Переменные окружения и shell",
		models.CategoryLogs:            "Логи и журналы системы",
		models.CategoryNetwork:         "Сеть, порты и соединения",
		models.CategoryPackages:        "Пакеты и менеджеры пакетов",
		models.CategoryProcesses:       "Процессы и нагрузка",
		models.CategoryServices:        "Сервисы и systemd",
		models.CategorySystem:          "Система, диски и ресурсы",
		models.CategoryText:            "Поиск и обработка текста",
		models.CategoryTroubleshooting: "Разбор сбоев, диагностика и поиск первопричины",
		models.CategoryUsers:           "Пользователи и сессии",
		"":                             "Команды категории",
	},
	"ua": {
		models.CategoryFilesystem:      "Файли, каталоги та права",
		models.CategoryEnvironment:     "Змінні середовища та shell",
		models.CategoryLogs:            "Логи та журнали системи",
		models.CategoryNetwork:         "Мережа, порти та з'єднання",
		models.CategoryPackages:        "Пакунки та менеджери пакунків",
		models.CategoryProcesses:       "Процеси та навантаження",
		models.CategoryServices:        "Сервіси та systemd",
		models.CategorySystem:          "Система, диски та ресурси",
		models.CategoryText:            "Пошук і обробка тексту",
		models.CategoryTroubleshooting: "Розбір збоїв, діагностика та пошук першопричини",
		models.CategoryUsers:           "Користувачі та сесії",
		"":                             "Команди категорії",
	},
}

const maxRecentVisible = 5
