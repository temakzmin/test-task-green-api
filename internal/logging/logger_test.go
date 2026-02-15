package logging

import (
	"testing"

	"github.com/stretchr/testify/require"

	"green-api/internal/config"
)

func TestNew_ValidConfig(t *testing.T) {
	t.Parallel()

	logger, err := New(config.LoggingConfig{Level: "info", Format: "json"})
	require.NoError(t, err)
	require.NotNil(t, logger)
}

func TestNew_InvalidLevel(t *testing.T) {
	t.Parallel()

	logger, err := New(config.LoggingConfig{Level: "verbose", Format: "json"})
	require.Error(t, err)
	require.Nil(t, logger)
}
