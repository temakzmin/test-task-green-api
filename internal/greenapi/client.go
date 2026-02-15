package greenapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"

	"green-api/internal/config"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	retry      config.GreenAPIRetryConfig
	breaker    *gobreaker.TwoStepCircuitBreaker
	logger     *zap.Logger
}

type Response struct {
	StatusCode  int
	Body        []byte
	ContentType string
	Headers     http.Header
}

type UpstreamError struct {
	Message string
	Cause   error
}

var ErrCircuitBreakerOpen = errors.New("green-api circuit breaker open")

func (e *UpstreamError) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *UpstreamError) Unwrap() error {
	return e.Cause
}

func NewClient(cfg config.GreenAPIConfig, logger *zap.Logger) *Client {
	settings := gobreaker.Settings{
		Name:        cfg.CircuitBreaker.Name,
		MaxRequests: cfg.CircuitBreaker.HalfOpenMaxRequests,
		Interval:    cfg.CircuitBreaker.Interval(),
		Timeout:     cfg.CircuitBreaker.OpenTimeout(),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < cfg.CircuitBreaker.MinRequests {
				return false
			}
			if counts.ConsecutiveFailures >= cfg.CircuitBreaker.ConsecutiveFailures {
				return true
			}
			if counts.Requests == 0 {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= cfg.CircuitBreaker.FailureRatio
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn("green_api_circuit_breaker_state_changed",
				zap.String("breaker", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	}

	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout()},
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		retry:      cfg.Retry,
		breaker:    gobreaker.NewTwoStepCircuitBreaker(settings),
		logger:     logger,
	}
}

func (c *Client) GetSettings(ctx context.Context, idInstance, apiTokenInstance string) (Response, error) {
	path := fmt.Sprintf("/waInstance%s/getSettings/%s", idInstance, apiTokenInstance)
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) GetStateInstance(ctx context.Context, idInstance, apiTokenInstance string) (Response, error) {
	path := fmt.Sprintf("/waInstance%s/getStateInstance/%s", idInstance, apiTokenInstance)
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) SendMessage(ctx context.Context, idInstance, apiTokenInstance, chatID, message string) (Response, error) {
	path := fmt.Sprintf("/waInstance%s/sendMessage/%s", idInstance, apiTokenInstance)
	payload := map[string]string{
		"chatId":  chatID,
		"message": message,
	}
	return c.do(ctx, http.MethodPost, path, payload)
}

func (c *Client) SendFileByURL(ctx context.Context, idInstance, apiTokenInstance, chatID, urlFile, fileName string) (Response, error) {
	path := fmt.Sprintf("/waInstance%s/sendFileByUrl/%s", idInstance, apiTokenInstance)
	payload := map[string]string{
		"chatId":   chatID,
		"urlFile":  urlFile,
		"fileName": fileName,
	}
	return c.do(ctx, http.MethodPost, path, payload)
}

func (c *Client) do(ctx context.Context, method, path string, payload any) (Response, error) {
	fullURL := c.baseURL + path
	if _, err := url.ParseRequestURI(fullURL); err != nil {
		return Response{}, &UpstreamError{Message: "invalid upstream url", Cause: err}
	}

	var bodyBytes []byte
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return Response{}, fmt.Errorf("marshal payload: %w", err)
		}
		bodyBytes = encoded
	}

	constantBackOff := backoff.NewConstantBackOff(c.retry.Delay())
	maxAttempts := c.retry.MaxRetries + 1

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := c.executeOnce(ctx, method, fullURL, bodyBytes)
		if err != nil {
			if attempt < maxAttempts && shouldRetryError(err) {
				c.wait(ctx, constantBackOff.NextBackOff())
				continue
			}
			if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
				return Response{}, &UpstreamError{Message: "green-api circuit breaker is open", Cause: fmt.Errorf("%w: %v", ErrCircuitBreakerOpen, err)}
			}
			return Response{}, &UpstreamError{Message: "green-api request failed", Cause: err}
		}

		response, readErr := readResponse(resp)
		if readErr != nil {
			return Response{}, &UpstreamError{Message: "read green-api response", Cause: readErr}
		}

		if response.StatusCode >= http.StatusInternalServerError && attempt < maxAttempts {
			c.wait(ctx, constantBackOff.NextBackOff())
			continue
		}

		return response, nil
	}

	return Response{}, &UpstreamError{Message: "green-api request failed after retries"}
}

func (c *Client) executeOnce(ctx context.Context, method, fullURL string, body []byte) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	done, err := c.breaker.Allow()
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		done(false)
		return nil, err
	}

	done(response.StatusCode < http.StatusInternalServerError)
	return response, nil
}

func readResponse(resp *http.Response) (Response, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	return Response{
		StatusCode:  resp.StatusCode,
		Body:        body,
		ContentType: resp.Header.Get("Content-Type"),
		Headers:     resp.Header,
	}, nil
}

func shouldRetryError(err error) bool {
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		var nestedNetErr net.Error
		if errors.As(urlErr, &nestedNetErr) {
			return nestedNetErr.Timeout()
		}
	}

	return false
}

func (c *Client) wait(ctx context.Context, delay time.Duration) {
	if delay <= 0 {
		return
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
