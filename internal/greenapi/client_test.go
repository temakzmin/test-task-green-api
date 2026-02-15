package greenapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"green-api/internal/config"
)

func testConfig(baseURL string) config.GreenAPIConfig {
	return config.GreenAPIConfig{
		BaseURL:        baseURL,
		TimeoutSeconds: 2,
		Retry: config.GreenAPIRetryConfig{
			MaxRetries:   2,
			DelaySeconds: 1,
		},
		CircuitBreaker: config.CircuitBreakerConfig{
			Name:                "test-breaker",
			ConsecutiveFailures: 50,
			HalfOpenMaxRequests: 1,
			OpenTimeoutSeconds:  60,
			IntervalSeconds:     60,
			FailureRatio:        1,
			MinRequests:         100,
		},
	}
}

func TestClient_RetryOn5xxThenSuccess(t *testing.T) {
	t.Parallel()

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requests, 1)
		if count == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"temporary"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"state":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(testConfig(server.URL), zap.NewNop())
	resp, err := client.GetSettings(context.Background(), "1101000001", "token")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, int32(2), atomic.LoadInt32(&requests))
}

func TestClient_DoesNotRetryOn4xx(t *testing.T) {
	t.Parallel()

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	client := NewClient(testConfig(server.URL), zap.NewNop())
	resp, err := client.GetSettings(context.Background(), "1101000001", "token")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, int32(1), atomic.LoadInt32(&requests))
}

func TestClient_CircuitBreakerOpenAfterFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"upstream down"}`))
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	cfg.Retry.MaxRetries = 0
	cfg.CircuitBreaker.ConsecutiveFailures = 1
	cfg.CircuitBreaker.MinRequests = 1
	cfg.CircuitBreaker.OpenTimeoutSeconds = 300
	client := NewClient(cfg, zap.NewNop())

	_, err := client.GetSettings(context.Background(), "1101000001", "token")
	require.NoError(t, err)

	_, err = client.GetSettings(context.Background(), "1101000001", "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), ErrCircuitBreakerOpen.Error())
}

func TestClient_SendFileByURLPayload(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/waInstance1101000001/sendFileByUrl/token", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"idMessage":"abc"}`))
	}))
	defer server.Close()

	client := NewClient(testConfig(server.URL), zap.NewNop())
	resp, err := client.SendFileByURL(context.Background(), "1101000001", "token", "77771234567@c.us", "https://x/img.png", "img.png")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, `{"idMessage":"abc"}`, string(resp.Body))
}

func TestClient_InvalidBaseURL(t *testing.T) {
	t.Parallel()

	cfg := testConfig("://invalid")
	client := NewClient(cfg, zap.NewNop())

	_, err := client.GetSettings(context.Background(), "1101000001", "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid upstream url")
}

func TestClient_GetStateInstancePath(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/waInstance123/getStateInstance/token", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"instanceState":"authorized"}`))
	}))
	defer server.Close()

	client := NewClient(testConfig(server.URL), zap.NewNop())
	resp, err := client.GetStateInstance(context.Background(), "123", "token")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, `{"instanceState":"authorized"}`, string(resp.Body))
}

func TestClient_SendMessagePath(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/waInstance123/sendMessage/token", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"idMessage":"1"}`))
	}))
	defer server.Close()

	client := NewClient(testConfig(server.URL), zap.NewNop())
	resp, err := client.SendMessage(context.Background(), "123", "token", "77771234567@c.us", "hello")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, `{"idMessage":"1"}`, string(resp.Body))
}

func TestClient_RetryOnServer5xxExhausted(t *testing.T) {
	t.Parallel()

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requests, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"attempt":%d}`, count)))
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	cfg.Retry.MaxRetries = 1
	client := NewClient(cfg, zap.NewNop())

	resp, err := client.GetSettings(context.Background(), "123", "token")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, int32(2), atomic.LoadInt32(&requests))
}
