package screens

import (
	"fmt"
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
	categoryCounts   []categoryCount
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
	categoryLabel    string
	categoryAll      string
	helpText         string
	width            int
	height           int
}

type categoryCount struct {
	category models.Category
	count    int
}

// NewCatalogModel constructs the initial recipe catalog screen.
func NewCatalogModel(recipes []models.Recipe, locale string, styles uitheme.Styles, favorites []string, recent []string, title string, emptyText string, recentTitle string, recentEmpty string, categoryLabel string, categoryAll string, helpText string) CatalogModel {
	orderedRecipes := orderResults(recipes, favoriteSet(favorites))
	model := CatalogModel{
		locale:        locale,
		styles:        styles,
		favorites:     favoriteSet(favorites),
		recent:        append([]string(nil), recent...),
		allRecipes:    orderedRecipes,
		categories:    categoriesFromResults(orderedRecipes),
		title:         title,
		emptyText:     emptyText,
		recentTitle:   recentTitle,
		recentEmpty:   recentEmpty,
		categoryLabel: categoryLabel,
		categoryAll:   categoryAll,
		helpText:      helpText,
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
		case "left":
			m.selectPreviousCategory()
			return m, nil
		case "right":
			m.selectNextCategory()
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
				if len(m.categoryCounts) > 0 {
					selected := m.categoryCounts[m.selected].category
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
	content := []string{m.styles.Title.Render(m.title), m.renderCategoryFilters(), ""}
	if m.isEmpty() {
		content = append(content, m.styles.Muted.Render(m.emptyText))
	} else {
		content = append(content, m.renderResults())
	}

	content = append(content, "", m.styles.Accent.Render(m.recentTitle))
	if len(m.recent) == 0 {
		content = append(content, m.styles.Muted.Render(m.recentEmpty))
	} else {
		content = append(content, m.renderRecent())
	}
	content = append(content, "", m.styles.Muted.Render(m.helpText))

	return renderFrame(m.styles, m.width, content)
}

// renderResults renders the current result list.
func (m CatalogModel) renderResults() string {
	if m.selectedCategory == "" {
		return m.renderCategoryRows()
	}

	lines := make([]string, 0, len(m.results))
	for index, recipe := range m.results {
		line := "  " + favoriteMarker(m.favorites, recipe.ID) + " " + resolveRecipeText(m.locale, recipe.Title) + " [" + string(recipe.Category) + "]"
		if index == m.selected {
			line = m.styles.Selected.Render("> " + favoriteMarker(m.favorites, recipe.ID) + " " + resolveRecipeText(m.locale, recipe.Title) + " [" + string(recipe.Category) + "]")
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m CatalogModel) renderCategoryRows() string {
	lines := make([]string, 0, len(m.categoryCounts))
	for index, entry := range m.categoryCounts {
		line := fmt.Sprintf("  %s (%d)", entry.category.DisplayName(), entry.count)
		if index == m.selected {
			line = m.styles.Selected.Render(fmt.Sprintf("> %s (%d)", entry.category.DisplayName(), entry.count))
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m CatalogModel) renderCategoryFilters() string {
	parts := make([]string, 0, len(m.categories)+1)
	parts = append(parts, m.renderCategoryOption("", m.categoryAll))
	for _, category := range m.categories {
		parts = append(parts, m.renderCategoryOption(category, category.DisplayName()))
	}

	return m.styles.Accent.Render(m.categoryLabel) + " " + strings.Join(parts, "  ")
}

func (m CatalogModel) renderCategoryOption(category models.Category, label string) string {
	if m.selectedCategory == category {
		return m.styles.Selected.Render("[" + label + "]")
	}

	return "[" + label + "]"
}

func (m CatalogModel) renderRecent() string {
	lines := make([]string, 0, min(maxRecentVisible, len(m.recent)))
	for index, command := range m.recent {
		if index >= maxRecentVisible {
			break
		}

		lines = append(lines, "- "+command)
	}

	return strings.Join(lines, "\n")
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

func (m *CatalogModel) selectPreviousCategory() {
	if len(m.categories) == 0 {
		return
	}

	selectedIndex := m.selectedCategoryIndex()
	if selectedIndex == 0 {
		m.selectedCategory = m.categories[len(m.categories)-1]
		m.applyCategoryFilter()
		return
	}

	if selectedIndex == -1 || len(m.categories) == 0 {
		return
	}

	if selectedIndex == 1 {
		m.selectedCategory = ""
		m.applyCategoryFilter()
		return
	}

	m.selectedCategory = m.categories[selectedIndex-2]
	m.applyCategoryFilter()
}

func (m *CatalogModel) selectNextCategory() {
	selectedIndex := m.selectedCategoryIndex()
	if len(m.categories) == 0 {
		return
	}

	if selectedIndex == -1 {
		return
	}

	if selectedIndex == len(m.categories) {
		m.selectedCategory = ""
		m.applyCategoryFilter()
		return
	}

	m.selectedCategory = m.categories[selectedIndex]
	m.applyCategoryFilter()
}

func (m CatalogModel) selectedCategoryIndex() int {
	if m.selectedCategory == "" {
		return 0
	}

	for index, category := range m.categories {
		if category == m.selectedCategory {
			return index + 1
		}
	}

	return -1
}

func (m *CatalogModel) applyCategoryFilter() {
	m.categories = categoriesFromResults(m.allRecipes)
	m.categoryCounts = categoryCountsFromResults(m.allRecipes, m.categories)
	if m.selectedCategory != "" && !containsCategory(m.categories, m.selectedCategory) {
		m.selectedCategory = ""
	}

	filtered := make([]models.Recipe, 0, len(m.allRecipes))
	if m.selectedCategory == "" {
		m.results = filtered
		if m.selected >= len(m.categoryCounts) {
			m.selected = max(0, len(m.categoryCounts)-1)
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
		return len(m.categoryCounts) == 0
	}

	return len(m.results) == 0
}

func (m CatalogModel) lastSelectableIndex() int {
	if m.selectedCategory == "" {
		return len(m.categoryCounts) - 1
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

func categoryCountsFromResults(results []models.Recipe, categories []models.Category) []categoryCount {
	counts := make([]categoryCount, 0, len(categories))
	for _, category := range categories {
		count := 0
		for _, recipe := range results {
			if recipe.Category == category {
				count++
			}
		}

		if count > 0 {
			counts = append(counts, categoryCount{category: category, count: count})
		}
	}

	return counts
}

func containsCategory(categories []models.Category, target models.Category) bool {
	for _, category := range categories {
		if category == target {
			return true
		}
	}

	return false
}

const maxRecentVisible = 5
