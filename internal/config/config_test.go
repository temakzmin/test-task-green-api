package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_ValidConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(cfgPath, []byte(`
server:
  host: 0.0.0.0
  port: 8080
  read_timeout_seconds: 15
  write_timeout_seconds: 15
  shutdown_timeout_seconds: 10
cors:
  allowed_origins:
    - http://localhost:5000
green_api:
  base_url: https://api.green-api.com
  timeout_seconds: 15
  retry:
    max_retries: 2
    delay_seconds: 1
  circuit_breaker:
    name: green-api
    consecutive_failures: 5
    half_open_max_requests: 1
    open_timeout_seconds: 30
    interval_seconds: 60
    failure_ratio: 0.5
    min_requests: 5
logging:
  level: info
  format: json
`), 0o644)
	require.NoError(t, err)

	cfg, err := Load(cfgPath)
	require.NoError(t, err)
	require.Equal(t, "0.0.0.0", cfg.Server.Host)
	require.Equal(t, 8080, cfg.Server.Port)
	require.Equal(t, "https://api.green-api.com", cfg.GreenAPI.BaseURL)
}

func TestLoad_InvalidConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.yaml")
	err := os.WriteFile(cfgPath, []byte(`
server:
  host: 0.0.0.0
`), 0o644)
	require.NoError(t, err)

	_, err = Load(cfgPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "validate config")
}
