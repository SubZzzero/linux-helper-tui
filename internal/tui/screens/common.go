package screens

import (
	"strings"
	"unicode"
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

// wrapPrefixedText splits plain text into width-limited lines with prefixes.
func wrapPrefixedText(value string, width int, firstPrefix string, continuationPrefix string) []string {
	if width <= 0 {
		return []string{firstPrefix + value}
	}

	remaining := strings.TrimSpace(value)
	lines := make([]string, 0, 2)
	prefix := firstPrefix

	for {
		available := max(1, width-textWidth(prefix))
		if textWidth(remaining) <= available {
			lines = append(lines, prefix+remaining)
			return lines
		}

		chunk := takeWrappedChunk(remaining, available)
		if chunk == "" {
			lines = append(lines, truncateText(prefix+remaining, width))
			return lines
		}

		lines = append(lines, prefix+chunk)
		remaining = strings.TrimSpace(strings.TrimPrefix(remaining, chunk))
		prefix = continuationPrefix
	}
}

func takeWrappedChunk(value string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= width {
		return string(runes)
	}

	cut := width
	for index := width - 1; index >= 0; index-- {
		if unicode.IsSpace(runes[index]) {
			cut = index
			break
		}
	}

	chunk := strings.TrimSpace(string(runes[:cut]))
	if chunk != "" {
		return chunk
	}

	return string(runes[:width])
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
