package i18n

import "linux-helper/internal/models"

// Translator resolves locale keys with fallback to English.
type Translator struct {
	locale string
	text   map[string]models.LocalizedText
}

// NewTranslator constructs a translator for one locale.
func NewTranslator(locale string, text map[string]models.LocalizedText) *Translator {
	return &Translator{locale: locale, text: text}
}

// T returns the translated string for one key.
func (t *Translator) T(key string) string {
	if text, ok := t.text[key]; ok {
		return text.Resolve(t.locale)
	}

	return key
}
