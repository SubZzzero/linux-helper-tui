package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"linux-helper/internal/models"
	uitheme "linux-helper/internal/tui/theme"
)

// FormModel collects recipe field values before execution.
type FormModel struct {
	recipe      models.Recipe
	locale      string
	styles      uitheme.Styles
	inputs      []textinput.Model
	selected    int
	submitText  string
	backText    string
	previewText string
	pendingBack bool
	pendingRun  map[string]string
	err         error
	width       int
}

// NewFormModel constructs one field-entry screen for a recipe.
func NewFormModel(recipe models.Recipe, locale string, styles uitheme.Styles, previewText string, submitText string, backText string) FormModel {
	inputs := make([]textinput.Model, 0, len(recipe.Fields))
	for index, field := range recipe.Fields {
		input := textinput.New()
		input.Prompt = ""
		input.Placeholder = field.Default
		input.SetValue(field.Default)
		if index == 0 {
			input.Focus()
		}
		inputs = append(inputs, input)
	}

	return FormModel{
		recipe:      recipe,
		locale:      locale,
		styles:      styles,
		inputs:      inputs,
		submitText:  submitText,
		backText:    backText,
		previewText: previewText,
	}
}

// Init starts the form screen with no async work.
func (m FormModel) Init() tea.Cmd {
	return nil
}

// Update handles text input, focus movement, and submission.
func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
	case tea.KeyMsg:
		switch typed.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.pendingBack = true
			return m, nil
		case "up", "shift+tab":
			m.focusPrevious()
			return m, nil
		case "down", "tab":
			m.focusNext()
			return m, nil
		case "enter":
			if len(m.inputs) == 0 || m.selected == len(m.inputs)-1 {
				values := m.values()
				if err := m.validateValues(values); err != nil {
					m.err = err
					return m, nil
				}
				m.err = nil
				m.pendingRun = values
				return m, nil
			}
			m.focusNext()
			return m, nil
		}
	}

	if len(m.inputs) == 0 {
		return m, nil
	}

	var cmd tea.Cmd
	m.inputs[m.selected], cmd = m.inputs[m.selected].Update(msg)
	return m, cmd
}

// ConsumeBack reports whether the form requested a pop.
func (m *FormModel) ConsumeBack() bool {
	back := m.pendingBack
	m.pendingBack = false
	return back
}

// ConsumeSubmit returns collected field values once.
func (m *FormModel) ConsumeSubmit() (map[string]string, bool) {
	if m.pendingRun == nil {
		return nil, false
	}

	values := m.pendingRun
	m.pendingRun = nil
	return values, true
}

// Recipe returns the recipe currently being filled.
func (m FormModel) Recipe() models.Recipe {
	return m.recipe
}

// Preview returns the rendered command summary for current field values.
func (m FormModel) Preview() string {
	return m.previewCommand()
}

// View renders the recipe field form and a live command summary.
func (m FormModel) View() string {
	lines := []string{
		m.styles.Title.Render(resolveRecipeText(m.locale, m.recipe.Title)),
		"",
		resolveRecipeText(m.locale, m.recipe.Description),
		"",
		m.styles.Accent.Render("Fields:"),
	}

	if len(m.recipe.Fields) == 0 {
		lines = append(lines, m.styles.Muted.Render("This recipe does not require additional fields."))
	} else {
		lines = append(lines, m.renderFields()...)
	}

	if m.err != nil {
		lines = append(lines, "", m.styles.Error.Render("Error: "+m.err.Error()))
	}

	lines = append(lines,
		"",
		m.styles.Accent.Render(m.previewText),
		m.previewCommand(),
		"",
		m.styles.Accent.Render(m.submitText),
		m.styles.Muted.Render(m.backText),
	)

	return renderFrame(m.styles, m.width, lines)
}

func (m *FormModel) focusPrevious() {
	if len(m.inputs) == 0 || m.selected == 0 {
		return
	}

	m.inputs[m.selected].Blur()
	m.selected--
	m.inputs[m.selected].Focus()
}

func (m *FormModel) focusNext() {
	if len(m.inputs) == 0 || m.selected >= len(m.inputs)-1 {
		return
	}

	m.inputs[m.selected].Blur()
	m.selected++
	m.inputs[m.selected].Focus()
}

func (m FormModel) renderFields() []string {
	lines := make([]string, 0, len(m.recipe.Fields)*2)
	for index, field := range m.recipe.Fields {
		label := fmt.Sprintf("%s:", field.Name)
		if field.Required {
			label += " *"
		}

		if description := resolveRecipeText(m.locale, field.Description); description != "" {
			label += " " + description
		}

		lines = append(lines, label)
		value := m.inputs[index].View()
		if index == m.selected {
			value = m.styles.Selected.Render(value)
		}
		lines = append(lines, value)
	}

	return lines
}

func (m FormModel) values() map[string]string {
	values := make(map[string]string, len(m.recipe.Fields))
	for index, field := range m.recipe.Fields {
		value := strings.TrimSpace(m.inputs[index].Value())
		if value == "" {
			value = field.Default
		}
		values[field.Name] = value
	}

	return values
}

func (m FormModel) previewCommand() string {
	values := m.values()
	parts := make([]string, 0, 4)
	if m.recipe.Execution == models.ExecutionTypeDirect {
		parts = append(parts, m.recipe.Binary)
		for _, arg := range m.recipe.Args {
			parts = append(parts, fillTemplate(arg, values))
		}
		return strings.Join(parts, " ")
	}

	return fillTemplate(m.recipe.Command, values)
}

func (m FormModel) validateValues(values map[string]string) error {
	for _, field := range m.recipe.Fields {
		if field.Required && strings.TrimSpace(values[field.Name]) == "" {
			return fmt.Errorf("field %q is required", field.Name)
		}
	}

	return nil
}

func fillTemplate(template string, values map[string]string) string {
	result := template
	for key, value := range values {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}

	return result
}
