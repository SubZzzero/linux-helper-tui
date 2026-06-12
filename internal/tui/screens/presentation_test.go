package screens_test

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"linux-helper/internal/models"
	"linux-helper/internal/tui/screens"
	uitheme "linux-helper/internal/tui/theme"
)

func altStyles() uitheme.Styles {
	return uitheme.NewStyles(uitheme.Definition{Name: "alt", BorderColor: "15", AccentColor: "10"})
}

// TestFormModelSetPresentationPreservesTypedValue keeps form state across locale refreshes.
func TestFormModelSetPresentationPreservesTypedValue(t *testing.T) {
	model := screens.NewFormModel(models.Recipe{
		Title:       models.LocalizedText{"en": "Find file"},
		Description: models.LocalizedText{"en": "Find files", "ua": "Знайти файли"},
		Execution:   models.ExecutionTypeDirect,
		Binary:      "find",
		Args:        []string{"{{path}}"},
		Fields: []models.Field{{
			Name:     "path",
			Type:     models.FieldTypeString,
			Required: true,
		}},
	}, "en", testStyles(), "Preview", "Submit", "Back")

	updated, _ := model.Update(tea.KeyMsg{Runes: []rune{'/'}, Type: tea.KeyRunes})
	formModel := updated.(screens.FormModel)
	formModel.SetPresentation("ua", altStyles(), "Попередній перегляд", "Надіслати", "Назад")

	view := formModel.View()
	assert.Contains(t, view, "Попередній перегляд")
	assert.Contains(t, view, "find /")
	assert.Contains(t, view, "Знайти файли")
}

// TestResultModelSetPresentationPreservesViewport keeps scrolled output visible.
func TestResultModelSetPresentationPreservesViewport(t *testing.T) {
	model := screens.NewResultModel(
		models.Recipe{Title: models.LocalizedText{"en": "Process list"}},
		"en",
		testStyles(),
		"Running",
		"Done",
		"Back",
		"Scroll",
	)

	lines := make([]string, 0, 20)
	for index := 1; index <= 20; index++ {
		lines = append(lines, fmt.Sprintf("line %02d", index))
	}

	model.SetOutcome(models.ExecutionResult{Command: "ps aux", ExitCode: 0, Stdout: strings.Join(lines, "\n")}, nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 16})
	resultModel := updated.(screens.ResultModel)

	for step := 0; step < 5; step++ {
		updated, _ = resultModel.Update(tea.KeyMsg{Type: tea.KeyPgDown})
		resultModel = updated.(screens.ResultModel)
	}

	require.Contains(t, resultModel.View(), "line 20")
	resultModel.SetPresentation("ua", altStyles(), "Виконується", "Готово", "Назад", "Прокрутка")
	view := resultModel.View()
	assert.Contains(t, view, "Прокрутка")
	assert.Contains(t, view, "line 20")
}
