package i18n_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"linux-helper/internal/i18n"
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
