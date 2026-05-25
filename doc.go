// Package jules provides a typed Go client for the Jules REST API.
//
// The package is the only public import path for this module. It exposes an
// opaque [Client], typed request and response models, resource helpers, artifact
// helpers, and session monitoring primitives for the documented Jules v1alpha
// resources: sources, sessions, session activities, and embedded artifacts.
//
// Construct a client with [NewClient]. The client owns authentication,
// transport retries, logging, and service wiring. Its configuration is immutable
// after construction from the caller's perspective; inspect the effective values
// with [Client.Config] and access API areas through service accessor methods:
//
//	client := jules.NewClient(
//		apiKey,
//		jules.WithBaseURL("https://jules.googleapis.com/v1alpha"),
//		jules.WithRetryAttempts(3),
//	)
//	sessions, err := client.Sessions().ListAll(ctx, 30, "")
//
// [Client.Sessions] creates, lists, archives, deletes, and messages sessions.
// [Client.Sources] lists and retrieves connected source repositories.
// [Client.Activities] lists session activity records and provides client-side
// filtering/search over documented activity payloads. [Client.Artifacts] reads
// embedded artifact data from activities and sessions.
//
// Source-backed session creation accepts either bare source IDs such as
// "github/owner/repo" or full resource names such as "sources/github/owner/repo".
// Bare session and source identifiers can also be normalized explicitly with
// [NormalizeSessionName] and [NormalizeSourceName].
//
// HTTP failures are returned as [*APIError] through normal Go error wrapping.
// The client retries transport errors, HTTP 429 responses, and HTTP 5xx
// responses. Retry-After is honored for rate limits, and debug logging redacts
// sensitive query parameters and authentication-like error fragments.
//
// [NewSessionMonitor] polls session state until completion, failure, user
// action, context cancellation, or timeout. Monitoring is intentionally a client
// helper only; the module does not include a CLI, config loader, patch applier,
// GitHub client, artifact writer, or local filesystem orchestration.
//
// API coverage is checked against Google's Jules Discovery document by the
// internal maintainer command:
//
//	go run ./internal/cmd/jules-api-coverage
//
// Importing applications should use only this package.
package jules
