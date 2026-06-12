package screens

import (
	"strings"
	"unicode/utf8"

	"linux-helper/internal/models"
	uitheme "linux-helper/internal/tui/theme"
)

const framedVerticalOverhead = 4
const framedHorizontalOverhead = 4

// renderFrame renders one framed screen with optional width expansion.
func renderFrame(styles uitheme.Styles, width int, lines []string) string {
	frame := styles.Frame
	if width > 0 {
		frame = frame.Width(max(1, width-2))
	}

	return frame.Render(strings.Join(lines, "\n"))
}

// textWidth returns the visible rune width for simple UI alignment.
func textWidth(value string) int {
	return utf8.RuneCountInString(value)
}

// truncateText cuts one plain-text line to the visible width budget.
func truncateText(value string, width int) string {
	if width <= 0 || textWidth(value) <= width {
		return value
	}

	if width <= 3 {
		runes := []rune(value)
		return string(runes[:width])
	}

	runes := []rune(value)
	return string(runes[:width-3]) + "..."
}

// resolveRecipeText resolves localized recipe text with English fallback.
func resolveRecipeText(locale string, text models.LocalizedText) string {
	return text.Resolve(locale)
}

// max returns the larger integer.
func max(left int, right int) int {
	if left > right {
		return left
	}

	return right
}

// min returns the smaller integer.
func min(left int, right int) int {
	if left < right {
		return left
	}

	return right
}
