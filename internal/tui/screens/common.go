package screens

import (
	"strings"

	"linux-helper/internal/models"
	uitheme "linux-helper/internal/tui/theme"
)

const minimumScreenWidth = 48

// renderFrame renders one framed screen with optional width expansion.
func renderFrame(styles uitheme.Styles, width int, lines []string) string {
	frame := styles.Frame
	if width > 0 {
		frame = frame.Width(max(minimumScreenWidth, width-2))
	}

	return frame.Render(strings.Join(lines, "\n"))
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
