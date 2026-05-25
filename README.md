# go-jules

`go-jules` is a reusable Go SDK for the Jules REST API.

The package exposes typed clients, focused resource services, requests,
responses, resource helpers, artifact helpers, and session monitoring
primitives for the documented Jules API v1alpha resources.

The SDK is intentionally a single importable package. The public package is a
thin facade over internal transport, model, resource, and service owners. It
does not include a CLI, config loader, patch applier, artifact writer, GitHub
client, or local filesystem side effects. Application orchestration belongs in
callers.

## Install

```bash
go get github.com/SamyRai/go-jules@latest
```

`go-jules` currently requires Go 1.25 or newer. CI verifies Go 1.25 and Go 1.26.

## Documentation

- [API guide](docs/api.md): client construction, service methods, resource
  names, activities, artifacts, monitoring, retries, and errors.
- [Development guide](docs/development.md): repository structure, ownership
  boundaries, verification commands, and API coverage checks.
- [Release notes and process](docs/releases.md): compatibility policy,
  unreleased changes, and tagging workflow.
- [Security policy](SECURITY.md): vulnerability reporting and runtime security
  considerations.

## Quick Start

```go
package main

import (
	"context"
	"log"
	"time"

	jules "github.com/SamyRai/go-jules"
)

func main() {
	client := jules.NewClient(
		"api-key",
		jules.WithBaseURL("https://jules.googleapis.com/v1alpha"),
		jules.WithTimeout(30*time.Second),
		jules.WithRetryAttempts(3),
	)

	sessions, err := client.Sessions().ListAll(context.Background(), 30, "")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("found %d sessions", len(sessions))
}
```

## Sessions

Create a repoless session:

```go
session, err := client.Sessions().Create(ctx, &jules.CreateSessionRequest{
	Prompt: "Explain the package architecture and suggest tests",
	Title:  "Architecture review",
})
if err != nil {
	return err
}
log.Println(session.Name, session.State)
```

Create a source-backed session. Bare source IDs are normalized to
`sources/{source}`. If `StartingBranch` is omitted, the SDK asks the Sources API
for the repository default branch.

```go
session, err := client.Sessions().Create(ctx, &jules.CreateSessionRequest{
	Prompt: "Fix the failing tests",
	SourceContext: &jules.SourceContext{
		Source: "github/owner/repo",
		GithubRepoContext: &jules.GithubRepoContext{
			StartingBranch: "main",
		},
	},
	RequirePlanApproval: true,
})
```

Send a follow-up message or approve a plan:

```go
err := client.Sessions().SendMessage(ctx, session.ID, &jules.SendMessageRequest{
	Prompt: "Also update the README example",
})

err = client.Sessions().ApprovePlan(ctx, session.ID)
```

## Activities and Artifacts

List activities with pagination or fetch only newer immutable activity records:

```go
cursor := time.Date(2026, 1, 17, 0, 3, 53, 137240000, time.UTC)
activities, err := client.Activities().ListSince(ctx, session.ID, cursor, 50)
if err != nil {
	return err
}
nextCursor := jules.ActivityCursor(activities)
```

Read embedded artifacts from documented activity payloads:

```go
artifacts, err := client.Artifacts().ListFromSession(ctx, session.ID)
if err != nil {
	return err
}
for _, item := range artifacts {
	content, err := jules.ArtifactContent(item.Artifact)
	if err != nil {
		continue
	}
	log.Printf("activity=%s artifact=%d bytes=%d", item.ActivityID, item.Index, len(content))
}
```

`ArtifactContent` returns git patch bytes for `changeSet.gitPatch`, formatted
command output for `bashOutput`, and decoded bytes for base64 `media`.
`ArtifactKindOf` and `ArtifactMetadataOf` classify artifacts without decoding
or writing their content.

## Monitoring

`SessionMonitor` checks the current session immediately, then polls until the
session completes, fails, needs user action, is cancelled by context, or reaches
the configured timeout. Intervals and max waits must be positive durations.

```go
status, err := jules.NewSessionMonitor(client, session.ID).
	WithInterval(5*time.Second).
	WithMaxWait(30*time.Minute).
	WaitForCompletion(ctx)
if err != nil {
	return err
}
if status.NeedsUserAction {
	log.Printf("session %s needs user action", status.Session.Name)
}
```

## Errors, Retries, and Logging

HTTP errors are returned as `*jules.APIError` through normal Go error wrapping:

```go
session, err := client.Sessions().Get(ctx, "sessions/missing")
if err != nil {
	var apiErr *jules.APIError
	if errors.As(err, &apiErr) && apiErr.IsNotFound() {
		return nil
	}
	return err
}
_ = session
```

The client retries transport errors, `429 Too Many Requests`, and 5xx responses.
`Retry-After` is honored for rate limits. Configure retries and structured debug
logging with client options:

```go
client := jules.NewClient(
	apiKey,
	jules.WithRetryAttempts(3),
	jules.WithRetryBackoff(time.Second),
	jules.WithLogger(logger),
	jules.WithDebugLog(true),
)
```

Debug logs redact sensitive query parameters and error details containing API
keys, tokens, auth values, secrets, or credentials.

## API Surface

- `client.Sources()`: list and get connected source repositories.
- `client.Sessions()`: create, list, get, delete, archive, unarchive, approve
  plans, and send messages.
- `client.Activities()`: list, get, cursor helpers, and client-side
  filtering/search over documented activity payloads.
- `client.Artifacts()`: read embedded `changeSet`, `bashOutput`, and base64
  `media` payloads from activity responses.
- Monitoring: poll session state until completion, user action, or timeout.

## Compatibility

The module is pre-1.0, but current consumers are treated as real. Breaking
changes should be explicit in release notes and reserved for API corrections or
architecture changes that materially improve maintainability.

The supported public import path is `github.com/SamyRai/go-jules`. Applications
should not import internal packages. API key loading, credential storage,
artifact writing, patch application, and repository orchestration are caller
responsibilities.

Versioning follows Go module conventions:

- patch releases fix bugs without public API changes
- minor releases may add API surface or carry documented pre-1.0 breaking
  changes
- tagged versions are immutable; fixes are published as new tags

See [release notes and process](docs/releases.md) for unreleased changes and
the release checklist. See [development guide](docs/development.md) for local
verification commands.

## License

MIT. See [LICENSE](LICENSE).
