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

## Monitoring

`SessionMonitor` polls until the session completes, fails, needs user action, is
cancelled by context, or reaches the configured timeout.

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

Versioning follows Go module conventions:

- patch releases fix bugs without public API changes
- minor releases may add API surface or carry documented pre-1.0 breaking
  changes
- tagged versions are immutable; fixes are published as new tags

The next planned release is `v0.2.0`.

## Release Notes

### Unreleased

- Introduce an opaque client facade with service accessors:
  `client.Sessions()`, `client.Sources()`, `client.Activities()`, and
  `client.Artifacts()`.
- Move DTOs, resource-name handling, HTTP transport/retry behavior, and service
  implementations behind internal ownership packages while keeping
  `github.com/SamyRai/go-jules` as the only public import path.
- Keep `go 1.25.0` as the minimum module version while testing Go 1.25 and Go
  1.26 in CI.
- Verify releases with Go 1.26.3.
- Send the documented activity `createTime` cursor query when
  `ListActivitiesOptions.CreateTime` is set, while keeping defensive
  client-side filtering.
- Add archive/unarchive session support and discovery-based API coverage
  validation.
- Harden retry, pagination, resource escaping, and option-default coverage.

## Development

```bash
gofmt -s -l .
go mod tidy
go test ./...
go test -race ./...
go vet ./...
go run ./internal/cmd/jules-api-coverage
GOTOOLCHAIN=go1.26.3 go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## Release Process

Releases are Go module releases. The package is published by pushing an
immutable semver git tag such as `v0.2.0`; consumers install it with:

```bash
go get github.com/SamyRai/go-jules@v0.2.0
```

Before tagging:

```bash
go mod tidy
go test ./...
go test -race ./...
go vet ./...
go run ./internal/cmd/jules-api-coverage
GOTOOLCHAIN=go1.26.3 go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Tag and publish from a clean main branch:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

The release workflow verifies formatting, module tidiness, tests, race tests,
`go vet`, `govulncheck`, Jules API coverage, and the module path. It then
creates a GitHub Release with generated notes and asks `proxy.golang.org` to
index the tag.

This SDK does not publish to GitHub Packages and does not build release assets:
Go modules are resolved from git tags through the Go module proxy. If a future
release adds CLI binaries or archives, attach them to the GitHub Release with
checksums, for example:

```bash
gh release upload vX.Y.Z dist/* checksums.txt --clobber
```

## License

MIT. See [LICENSE](LICENSE).
