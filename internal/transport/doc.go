// Package transport owns Jules HTTP request execution, retries, API errors, and
// debug logging.
//
// Transport is the infrastructure boundary below the SDK services. It marshals
// request bodies, attaches API-key authentication and JSON headers, executes
// requests with a configured http.Client, retries retryable failures, decodes
// successful JSON responses, and converts final non-2xx responses into
// model.APIError values.
//
// Retry behavior is intentionally limited to transport errors, HTTP 429, and
// HTTP 5xx responses. Retry-After is honored for rate limits; otherwise
// exponential backoff is based on Config.RetryBackoff. Request bodies are
// replayed from net/http's GetBody support, which is provided for buffers
// created by http.NewRequest.
//
// Debug logging is opt-in. When enabled, request logs include method, URL,
// status, duration, and attempt count. Query parameters and error fragments that
// look like API keys, tokens, auth values, secrets, or credentials are redacted
// before logging.
package transport
