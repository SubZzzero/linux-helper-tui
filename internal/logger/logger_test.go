package logger_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"linux-helper/internal/logger"
)

// TestNew creates a file-backed logger.
func TestNew(t *testing.T) {
	log, closer, err := logger.New(filepath.Join(t.TempDir(), "app.log"))
	require.NoError(t, err)
	require.NotNil(t, log)
	require.NoError(t, closer.Close())
}
