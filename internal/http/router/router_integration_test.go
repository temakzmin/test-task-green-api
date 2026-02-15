package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"green-api/internal/config"
	"green-api/internal/greenapi"
	"green-api/internal/service"
)

func integrationConfig(baseURL string) config.Config {
	return config.Config{
		Server: config.ServerConfig{
			Host:                   "127.0.0.1",
			Port:                   8080,
			ReadTimeoutSeconds:     5,
			WriteTimeoutSeconds:    5,
			ShutdownTimeoutSeconds: 5,
		},
		CORS: config.CORSConfig{AllowedOrigins: []string{"http://localhost:5000"}},
		GreenAPI: config.GreenAPIConfig{
			BaseURL:        baseURL,
			TimeoutSeconds: 5,
			Retry: config.GreenAPIRetryConfig{
				MaxRetries:   1,
				DelaySeconds: 1,
			},
			CircuitBreaker: config.CircuitBreakerConfig{
				Name:                "integration-breaker",
				ConsecutiveFailures: 5,
				HalfOpenMaxRequests: 1,
				OpenTimeoutSeconds:  30,
				IntervalSeconds:     30,
				FailureRatio:        0.5,
				MinRequests:         3,
			},
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

func TestRouter_GetSettingsProxy(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/waInstance1101000001/getSettings/token", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"wid":"79990000000"}`))
	}))
	defer upstream.Close()

	cfg := integrationConfig(upstream.URL)
	logger := zap.NewNop()
	client := greenapi.NewClient(cfg.GreenAPI, logger)
	svc := service.New(client)
	engine := New(cfg, logger, svc)

	body, _ := json.Marshal(map[string]string{
		"idInstance":       "1101000001",
		"apiTokenInstance": "token",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	engine.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)
	require.JSONEq(t, `{"wid":"79990000000"}`, resp.Body.String())
	require.NotEmpty(t, resp.Header().Get("X-Request-Id"))
}

func TestRouter_SendMessageValidationError(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"idMessage":"x"}`))
	}))
	defer upstream.Close()

	cfg := integrationConfig(upstream.URL)
	logger := zap.NewNop()
	client := greenapi.NewClient(cfg.GreenAPI, logger)
	svc := service.New(client)
	engine := New(cfg, logger, svc)

	body, _ := json.Marshal(map[string]string{
		"idInstance":       "1101000001",
		"apiTokenInstance": "token",
		"chatId":           "invalid",
		"message":          "hello",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/send-message", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	engine.ServeHTTP(resp, req)
	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Contains(t, resp.Body.String(), "validation_error")
}

func TestRouter_DocsEndpointsAvailable(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	cfg := integrationConfig(upstream.URL)
	logger := zap.NewNop()
	client := greenapi.NewClient(cfg.GreenAPI, logger)
	svc := service.New(client)
	engine := New(cfg, logger, svc)

	openapiReq := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	openapiResp := httptest.NewRecorder()
	engine.ServeHTTP(openapiResp, openapiReq)
	require.Equal(t, http.StatusOK, openapiResp.Code)
	require.Contains(t, openapiResp.Body.String(), "openapi: 3.0.3")

	docsReq := httptest.NewRequest(http.MethodGet, "/docs/index.html", nil)
	docsResp := httptest.NewRecorder()
	engine.ServeHTTP(docsResp, docsReq)
	require.Equal(t, http.StatusOK, docsResp.Code)
	require.Contains(t, docsResp.Body.String(), "SwaggerUIBundle")
}
