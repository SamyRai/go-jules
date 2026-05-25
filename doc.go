// Package jules provides a typed Go client for the Jules REST API.
//
// The package covers the documented Jules v1alpha resources: sources,
// sessions, session activities, embedded artifacts, and session monitoring.
// It is intentionally limited to API client behavior and does not include a
// CLI, config loader, patch applier, GitHub client, or local filesystem
// orchestration. The public package is a facade over internal model, resource,
// transport, and service owners.
//
// Create a client with NewClient and call focused resource services through
// Client.Sessions(), Client.Sources(), Client.Activities(), and Client.Artifacts().
// Pagination and official API filters are exposed through ListSessionsOptions,
// ListSourcesOptions, and ListActivitiesOptions.
//
// The client uses API-key authentication through the X-Goog-Api-Key header,
// retries transport errors, 429 responses, and 5xx responses, and returns HTTP
// failures as *APIError through normal Go error wrapping.
//
// API coverage is checked against Google's Jules Discovery document by the
// internal coverage command:
//
//	go run ./internal/cmd/jules-api-coverage
//
// That command is for this module's maintainers; importing applications should
// use only the public package API.
package jules
