package jules

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/SamyRai/go-jules/internal/services"
	"github.com/SamyRai/go-jules/internal/transport"
)

const defaultBaseURL = "https://jules.googleapis.com/v1alpha"
const defaultUserAgent = "juleson-go-sdk"

// SleepFunc sleeps for the provided duration and may return early with an error.
type SleepFunc = transport.SleepFunc

// ClientConfig contains the effective configuration for a Client.
type ClientConfig struct {
	APIKey        string
	BaseURL       string
	HTTPClient    *http.Client
	RetryAttempts int
	RetryBackoff  time.Duration
	UserAgent     string
	Sleep         SleepFunc
	Logger        *slog.Logger
	DebugLog      bool
}

// Client represents a Jules API client.
type Client struct {
	apiKey        string
	baseURL       string
	httpClient    *http.Client
	retryAttempts int
	retryBackoff  time.Duration
	userAgent     string
	sleep         SleepFunc
	logger        *slog.Logger
	debugLog      bool
	transport     *transport.Transport

	sessions   *SessionsService
	sources    *SourcesService
	activities *ActivitiesService
	artifacts  *ArtifactsService
}

// ClientOption configures a Jules API client.
type ClientOption func(*Client)

// NewClient creates a new Jules API client.
func NewClient(apiKey string, options ...ClientOption) *Client {
	client := &Client{
		apiKey:        apiKey,
		baseURL:       defaultBaseURL,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		retryAttempts: 3,
		retryBackoff:  time.Second,
		userAgent:     defaultUserAgent,
	}

	for _, option := range options {
		if option != nil {
			option(client)
		}
	}

	if client.httpClient == nil {
		client.httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	if client.baseURL == "" {
		client.baseURL = defaultBaseURL
	}
	client.baseURL = strings.TrimRight(client.baseURL, "/")
	if client.retryAttempts < 0 {
		client.retryAttempts = 0
	}
	if client.retryBackoff <= 0 {
		client.retryBackoff = time.Second
	}
	if client.userAgent == "" {
		client.userAgent = defaultUserAgent
	}
	if client.sleep == nil {
		client.sleep = func(d time.Duration) error {
			time.Sleep(d)
			return nil
		}
	}
	client.transport = transport.New(transport.Config{
		APIKey:        client.apiKey,
		HTTPClient:    client.httpClient,
		RetryAttempts: client.retryAttempts,
		RetryBackoff:  client.retryBackoff,
		UserAgent:     client.userAgent,
		Sleep:         client.sleep,
		Logger:        client.logger,
		DebugLog:      client.debugLog,
	})
	client.sources = services.NewSourcesService(client.baseURL, client.transport)
	client.sessions = services.NewSessionsService(client.baseURL, client.transport, client.sources)
	client.activities = services.NewActivitiesService(client.baseURL, client.transport)
	client.artifacts = services.NewArtifactsService(client.activities)

	return client
}

// WithBaseURL sets the Jules API base URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = timeout
	}
}

// WithRetryAttempts sets the number of retry attempts for retryable requests.
func WithRetryAttempts(retryAttempts int) ClientOption {
	return func(c *Client) {
		c.retryAttempts = retryAttempts
	}
}

// WithHTTPClient sets the HTTP client used for requests.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithRetryBackoff sets the base retry backoff duration.
func WithRetryBackoff(backoff time.Duration) ClientOption {
	return func(c *Client) {
		c.retryBackoff = backoff
	}
}

// WithUserAgent sets the User-Agent header.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithLogger sets the logger for the client.
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithDebugLog enables or disables debug logging.
func WithDebugLog(debugLog bool) ClientOption {
	return func(c *Client) {
		c.debugLog = debugLog
	}
}

// WithSleep sets the sleep function used between retries. It is primarily
// intended for tests.
func WithSleep(sleep SleepFunc) ClientOption {
	return func(c *Client) {
		c.sleep = sleep
	}
}

// Config returns the effective client configuration.
func (c *Client) Config() ClientConfig {
	return ClientConfig{
		APIKey:        c.apiKey,
		BaseURL:       c.baseURL,
		HTTPClient:    c.httpClient,
		RetryAttempts: c.retryAttempts,
		RetryBackoff:  c.retryBackoff,
		UserAgent:     c.userAgent,
		Sleep:         c.sleep,
		Logger:        c.logger,
		DebugLog:      c.debugLog,
	}
}

// Sessions returns the session API service.
func (c *Client) Sessions() *SessionsService {
	return c.sessions
}

// Sources returns the source API service.
func (c *Client) Sources() *SourcesService {
	return c.sources
}

// Activities returns the activity API service.
func (c *Client) Activities() *ActivitiesService {
	return c.activities
}

// Artifacts returns the artifact helper service.
func (c *Client) Artifacts() *ArtifactsService {
	return c.artifacts
}
