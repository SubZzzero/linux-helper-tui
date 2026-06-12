package i18n_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	linuxhelper "linux-helper"
	"linux-helper/internal/i18n"
	"linux-helper/internal/models"
)

// TestLoadLocales loads JSON locale catalogues.
func TestLoadLocales(t *testing.T) {
	fsys := fstest.MapFS{
		"locales/en.json": {Data: []byte(`{"hello":"Hello"}`)},
		"locales/ua.json": {Data: []byte(`{"hello":"Привіт"}`)},
	}

	translations, err := i18n.LoadLocales(fsys, "locales")
	require.NoError(t, err)

	translator := i18n.NewTranslator("ua", translations)
	assert.Equal(t, "Привіт", translator.T("hello"))
	assert.Equal(t, "missing", translator.T("missing"))
}

// TestLoadEmbeddedLocalesIncludesEnUaRu loads the bundled locale catalogues.
func TestLoadEmbeddedLocalesIncludesEnUaRu(t *testing.T) {
	localeFS, err := fs.Sub(linuxhelper.Assets, "assets/locales")
	require.NoError(t, err)

	translations, err := i18n.LoadLocales(localeFS, ".")
	require.NoError(t, err)
	assert.Equal(t, "Recent commands", i18n.NewTranslator("en", translations).T("catalog.recent_title"))
	assert.Equal(t, "Останні команди", i18n.NewTranslator("ua", translations).T("catalog.recent_title"))
	assert.Equal(t, "Последние команды", i18n.NewTranslator("ru", translations).T("catalog.recent_title"))
}

// TestTranslatorFallsBackToEnglishWhenLocaleEntryMissing uses the English value when needed.
func TestTranslatorFallsBackToEnglishWhenLocaleEntryMissing(t *testing.T) {
	translations := map[string]models.LocalizedText{
		"hello": {"en": "Hello"},
	}
	assert.Equal(t, "Hello", i18n.NewTranslator("ua", translations).T("hello"))
}
