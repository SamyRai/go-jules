# API Guide

`go-jules` exposes one public import path:

```go
import jules "github.com/SamyRai/go-jules"
```

The public package is a facade over internal model, resource, service, and
transport packages. Importing applications should use only
`github.com/SamyRai/go-jules`; internal packages are implementation details.

## Client Construction

Create a client with an API key and optional configuration:

```go
client := jules.NewClient(
	apiKey,
	jules.WithBaseURL("https://jules.googleapis.com/v1alpha"),
	jules.WithTimeout(30*time.Second),
	jules.WithRetryAttempts(3),
	jules.WithRetryBackoff(time.Second),
	jules.WithLogger(logger),
	jules.WithDebugLog(true),
)
```

Defaults:

- base URL: `https://jules.googleapis.com/v1alpha`
- HTTP timeout: `30s`
- retry attempts: `3`
- retry backoff: `1s`
- user agent: `juleson-go-sdk`

Use `client.Config()` to inspect the effective immutable configuration. Use
`WithHTTPClient` for a caller-owned `*http.Client` and `WithSleep` only when a
test needs deterministic retry timing.

## Services

Access Jules resources through service accessors:

- `client.Sources()`: list, list all, and get connected source repositories.
- `client.Sessions()`: create, list, list all, get, delete, archive,
  unarchive, approve plans, and send messages.
- `client.Activities()`: list, list all, list since a cursor, get, filter, and
  search documented activity payloads.
- `client.Artifacts()`: list embedded artifacts from an activity or session and
  read documented embedded content.

Pagination helpers named `ListAll` follow `nextPageToken` until the API returns
an empty token. Page sizes are normalized by the services and capped at `100`.

## Resource Names

The SDK accepts bare IDs for common calls and normalizes them to Jules resource
names:

- `jules.NormalizeSessionName("abc")` returns `sessions/abc`
- `jules.NormalizeSourceName("github/owner/repo")` returns
  `sources/github/owner/repo`

Source IDs may contain path segments such as `github/owner/repo`. Session and
activity IDs are escaped when used in request paths.

## Sessions

Repoless sessions require a prompt:

```go
session, err := client.Sessions().Create(ctx, &jules.CreateSessionRequest{
	Prompt: "Explain the package architecture and suggest tests",
	Title:  "Architecture review",
})
```

Source-backed sessions accept bare source IDs. If `StartingBranch` is omitted,
the SDK looks up the source and uses the default branch when the Sources API
returns one:

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

Plan and message helpers call the documented session actions:

```go
err := client.Sessions().SendMessage(ctx, session.ID, &jules.SendMessageRequest{
	Prompt: "Also update the README example",
})
err = client.Sessions().ApprovePlan(ctx, session.ID)
```

Archive state is changed with `Archive` and `Unarchive`; both return the
updated session.

## Activities and Artifacts

`ListActivitiesOptions.CreateTime` sends the documented `createTime` cursor
query and also filters returned activities client-side:

```go
activities, err := client.Activities().ListSince(ctx, session.ID, cursor, 50)
if err != nil {
	return err
}
cursor = jules.ActivityCursor(activities)
```

Client-side filtering and search work over documented activity fields and
payloads. They do not call an undocumented search endpoint.

Artifacts are embedded in activity responses:

```go
items, err := client.Artifacts().ListFromSession(ctx, session.ID)
if err != nil {
	return err
}
for _, item := range items {
	content, err := jules.ArtifactContent(item.Artifact)
	if err != nil {
		continue
	}
	_ = content
}
```

`ArtifactContent` returns:

- git patch bytes for `changeSet.gitPatch`
- formatted command output for `bashOutput`
- base64-decoded bytes for `media`

Use `ArtifactKindOf`, `ArtifactMetadataOf`, or the matching methods on
`client.Artifacts()` to classify artifacts without decoding content or writing
files.

## Monitoring

`SessionMonitor` checks the session immediately, then polls until completion,
failure, required user action, context cancellation, or timeout:

```go
status, err := jules.NewSessionMonitor(client, session.ID).
	WithInterval(5*time.Second).
	WithMaxWait(30*time.Minute).
	WaitForCompletion(ctx)
```

`WithInterval` and `WithMaxWait` must be positive durations. `WaitForCompletion`
and `PollUntilComplete` use continuous polling. `WaitForPlan` polls activities
until a `planGenerated` payload appears and returns activity-list errors instead
of hiding them until timeout.

## Errors and Retries

HTTP failures are returned as `*jules.APIError` through normal Go error
wrapping:

```go
var apiErr *jules.APIError
if errors.As(err, &apiErr) && apiErr.IsNotFound() {
	return nil
}
```

The transport retries network errors, `429 Too Many Requests`, and HTTP 5xx
responses. `Retry-After` is honored for rate limits. Non-429 HTTP 4xx responses
are not retried.

When debug logging is enabled with `WithDebugLog(true)` and a logger is set, the
transport emits structured request logs. Sensitive query parameters and error
details containing key, token, auth, secret, or credential fragments are
redacted.

## Non-Goals

The SDK does not include a CLI, config loader, patch applier, artifact writer,
GitHub client, credential store, or local filesystem orchestration. Those
concerns belong in applications that import this module.
