package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"green-api/internal/greenapi"
	"green-api/internal/service"
)

type mockClient struct{}

func (m *mockClient) GetSettings(context.Context, string, string) (greenapi.Response, error) {
	return greenapi.Response{StatusCode: http.StatusOK, Body: []byte(`{"wid":"79990000000"}`), ContentType: "application/json"}, nil
}

func (m *mockClient) GetStateInstance(context.Context, string, string) (greenapi.Response, error) {
	return greenapi.Response{StatusCode: http.StatusOK, Body: []byte(`{"state":"authorized"}`), ContentType: "application/json"}, nil
}

func (m *mockClient) SendMessage(context.Context, string, string, string, string) (greenapi.Response, error) {
	return greenapi.Response{StatusCode: http.StatusOK, Body: []byte(`{"idMessage":"1"}`), ContentType: "application/json"}, nil
}

func (m *mockClient) SendFileByURL(context.Context, string, string, string, string, string) (greenapi.Response, error) {
	return greenapi.Response{StatusCode: http.StatusOK, Body: []byte(`{"idMessage":"2"}`), ContentType: "application/json"}, nil
}

func setupHandlerRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := service.New(&mockClient{})
	h := NewGreenAPIHandler(svc)
	group := r.Group("/api/v1")
	h.RegisterRoutes(group)
	return r
}

func TestGetSettings_Success(t *testing.T) {
	t.Parallel()

	r := setupHandlerRouter()
	body := `{"idInstance":"1101000001","apiTokenInstance":"token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	require.JSONEq(t, `{"wid":"79990000000"}`, resp.Body.String())
}

func TestSendMessage_InvalidJSON(t *testing.T) {
	t.Parallel()

	r := setupHandlerRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/send-message", strings.NewReader(`{"idInstance":`))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Contains(t, resp.Body.String(), "bad_request")
}
