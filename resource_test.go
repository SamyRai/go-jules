package jules

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceNameHandling(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := NewClient("test-api-key", WithBaseURL("https://jules.googleapis.com/v1alpha"), WithRetryAttempts(0))

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/123",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(http.StatusOK, Session{ID: "123"})
		})
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sources/github/owner/repo",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(http.StatusOK, Source{ID: "github/owner/repo"})
		})
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sources/github/owner%20name/repo%20name",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(http.StatusOK, Source{ID: "github/owner name/repo name"})
		})
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/123/activities/act1",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(http.StatusOK, Activity{ID: "act1"})
		})

	session, err := client.Sessions.Get(context.Background(), "sessions/123")
	require.NoError(t, err)
	assert.Equal(t, "123", session.ID)

	source, err := client.Sources.Get(context.Background(), "sources/github/owner/repo")
	require.NoError(t, err)
	assert.Equal(t, "github/owner/repo", source.ID)

	source, err = client.Sources.Get(context.Background(), "sources/github/owner name/repo name")
	require.NoError(t, err)
	assert.Equal(t, "github/owner name/repo name", source.ID)

	activity, err := client.Activities.Get(context.Background(), "ignored", "sessions/123/activities/act1")
	require.NoError(t, err)
	assert.Equal(t, "act1", activity.ID)

	_, err = client.Sessions.Get(context.Background(), "sessions/a/b")
	require.Error(t, err)

	_, err = client.Sources.Get(context.Background(), "sources/github//repo")
	require.Error(t, err)
}

func TestRetryAfterAndAPIErrorDetails(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var slept []time.Duration
	client := NewClient(
		"secret-api-key",
		WithBaseURL("https://jules.googleapis.com/v1alpha"),
		WithRetryAttempts(1),
		WithSleep(func(d time.Duration) error {
			slept = append(slept, d)
			return nil
		}),
	)

	calls := 0
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/ratelimited",
		func(req *http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				resp := httpmock.NewStringResponse(http.StatusTooManyRequests, "rate limited")
				resp.Header.Set("Retry-After", "2")
				return resp, nil
			}
			return httpmock.NewStringResponse(http.StatusForbidden, "permission denied for request"), nil
		})

	_, err := client.Sessions.Get(context.Background(), "ratelimited")
	require.Error(t, err)
	assert.Equal(t, []time.Duration{2 * time.Second}, slept)
	assert.Equal(t, 2, calls)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
	assert.Equal(t, "GET", apiErr.Method)
	assert.Equal(t, "/v1alpha/sessions/ratelimited", apiErr.Path)
	assert.NotContains(t, apiErr.Error(), "secret-api-key")
}
