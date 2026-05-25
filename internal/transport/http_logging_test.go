package transport

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientLogging_DisabledByDefault(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/123",
		httpmock.NewStringResponder(200, `{}`))

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client := New(Config{
		APIKey:     "test-key",
		HTTPClient: &http.Client{},
		Logger:     logger,
	}) // Note: DebugLog is false by default

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://jules.googleapis.com/v1alpha/sessions/123", nil)
	_, err := client.do(req)
	require.NoError(t, err)

	assert.Empty(t, logBuf.String(), "Expected no logs to be written when debugLog is false")
}

func TestClientLogging_EnabledLogsRequestData(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/123",
		httpmock.NewStringResponder(200, `{}`))

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client := New(Config{
		APIKey:     "test-key",
		HTTPClient: &http.Client{},
		Logger:     logger,
		DebugLog:   true,
	})

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://jules.googleapis.com/v1alpha/sessions/123", nil)
	_, err := client.do(req)
	require.NoError(t, err)

	logOutput := logBuf.String()
	assert.NotEmpty(t, logOutput, "Expected logs to be written when debugLog is true")
	assert.Contains(t, logOutput, "method=GET")
	assert.Contains(t, logOutput, "url=https://jules.googleapis.com/v1alpha/sessions/123")
	assert.Contains(t, logOutput, "status_code=200")
	assert.Contains(t, logOutput, "attempt=1")
	assert.Contains(t, logOutput, "duration=")
}

func TestClientLogging_RedactsSensitiveQueryParams(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "=~^https://jules.googleapis.com/v1alpha/sessions.*",
		httpmock.NewStringResponder(200, `{}`))

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client := New(Config{
		APIKey:     "test-key",
		HTTPClient: &http.Client{},
		Logger:     logger,
		DebugLog:   true,
	})

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://jules.googleapis.com/v1alpha/sessions?api_key=secret1&token=secret2&auth_token=secret3&credential=secret4&safe_param=hello", nil)
	_, err := client.do(req)
	require.NoError(t, err)

	logOutput := logBuf.String()
	assert.NotContains(t, logOutput, "secret1", "Expected api_key to be redacted")
	assert.NotContains(t, logOutput, "secret2", "Expected token to be redacted")
	assert.NotContains(t, logOutput, "secret3", "Expected auth_token to be redacted")
	assert.NotContains(t, logOutput, "secret4", "Expected credential to be redacted")

	assert.Contains(t, logOutput, "api_key=REDACTED")
	assert.Contains(t, logOutput, "token=REDACTED")
	assert.Contains(t, logOutput, "auth_token=REDACTED")
	assert.Contains(t, logOutput, "credential=REDACTED")
	assert.Contains(t, logOutput, "safe_param=hello", "Expected non-sensitive params to remain unredacted")
}

func TestClientLogging_LogsRetriesAndErrors(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	callCount := 0
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/123",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount < 2 {
				return httpmock.NewStringResponse(500, "Internal Server Error"), nil
			}
			return httpmock.NewStringResponse(200, `{}`), nil
		})

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client := New(Config{
		APIKey:        "test-key",
		HTTPClient:    &http.Client{},
		Logger:        logger,
		DebugLog:      true,
		RetryAttempts: 2,
		RetryBackoff:  1 * time.Millisecond,
	})

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://jules.googleapis.com/v1alpha/sessions/123", nil)
	_, err := client.do(req)
	require.NoError(t, err)

	logOutput := logBuf.String()

	// Should see attempt 1 failing
	assert.Contains(t, logOutput, "attempt=1")
	assert.Contains(t, logOutput, "status_code=500")

	// Should see attempt 2 succeeding
	assert.Contains(t, logOutput, "attempt=2")
	assert.Contains(t, logOutput, "status_code=200")

	// Split lines to verify order
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	assert.Len(t, lines, 2)
}

func TestClientLogging_RedactsSensitiveErrorDetails(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client := New(Config{
		APIKey: "test-key",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New(`Get "https://jules.googleapis.com/v1alpha/sessions?api_key=secret1&token=secret2": failed`)
			}),
		},
		Logger:        logger,
		DebugLog:      true,
		RetryAttempts: 0,
	})

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://jules.googleapis.com/v1alpha/sessions?api_key=secret1&token=secret2", nil)
	_, err := client.do(req)
	require.Error(t, err)

	logOutput := logBuf.String()
	assert.NotContains(t, logOutput, "secret1")
	assert.NotContains(t, logOutput, "secret2")
	assert.Contains(t, logOutput, "api_key=REDACTED")
	assert.Contains(t, logOutput, "token=REDACTED")
}

func TestDoJSON_AllowsEmptySuccessfulResponseWithResult(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/empty",
		httpmock.NewStringResponder(http.StatusOK, ""))

	client := New(Config{
		APIKey:        "test-key",
		HTTPClient:    &http.Client{},
		RetryAttempts: 0,
	})

	var result struct {
		ID string `json:"id"`
	}
	err := client.DoJSON(context.Background(), "GET", "https://jules.googleapis.com/v1alpha/sessions/empty", nil, &result)

	require.NoError(t, err)
	assert.Empty(t, result.ID)
}

func TestDo_DoesNotRetryNonRateLimitedClientErrors(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	callCount := 0
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/forbidden",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			return httpmock.NewStringResponse(http.StatusForbidden, "permission denied"), nil
		})

	client := New(Config{
		APIKey:        "test-key",
		HTTPClient:    &http.Client{},
		RetryAttempts: 3,
		RetryBackoff:  time.Millisecond,
	})

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://jules.googleapis.com/v1alpha/sessions/forbidden", nil)
	_, err := client.do(req)

	require.Error(t, err)
	assert.Equal(t, 1, callCount)
}

func TestDo_HonorsRetryAfterHTTPDate(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var slept []time.Duration
	callCount := 0
	retryAt := time.Now().Add(time.Hour).UTC().Format(http.TimeFormat)
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/rate-limited",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				resp := httpmock.NewStringResponse(http.StatusTooManyRequests, "rate limited")
				resp.Header.Set("Retry-After", retryAt)
				return resp, nil
			}
			return httpmock.NewStringResponse(http.StatusOK, `{}`), nil
		})

	client := New(Config{
		APIKey:        "test-key",
		HTTPClient:    &http.Client{},
		RetryAttempts: 1,
		RetryBackoff:  time.Millisecond,
		Sleep: func(d time.Duration) error {
			slept = append(slept, d)
			return nil
		},
	})

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://jules.googleapis.com/v1alpha/sessions/rate-limited", nil)
	resp, err := client.do(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 2, callCount)
	require.Len(t, slept, 1)
	assert.Greater(t, slept[0], 59*time.Minute)
	assert.LessOrEqual(t, slept[0], time.Hour)
}

func TestDo_ReturnsContextErrorWhenCancelledDuringRetrySleep(t *testing.T) {
	sleepStarted := make(chan struct{})
	sleepRelease := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := New(Config{
		APIKey: "test-key",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("try again")),
					Header:     make(http.Header),
				}, nil
			}),
		},
		RetryAttempts: 1,
		RetryBackoff:  time.Hour,
		Sleep: func(time.Duration) error {
			close(sleepStarted)
			<-sleepRelease
			return nil
		},
	})

	errCh := make(chan error, 1)
	go func() {
		req, _ := http.NewRequestWithContext(ctx, "GET", "https://jules.googleapis.com/v1alpha/sessions/retry", nil)
		_, err := client.do(req)
		errCh <- err
	}()

	<-sleepStarted
	cancel()
	err := <-errCh
	close(sleepRelease)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
