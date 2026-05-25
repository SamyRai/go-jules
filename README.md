# go-jules

`go-jules` is a reusable Go SDK for the Jules REST API.

The package exposes typed clients, requests, responses, resource helpers,
artifact helpers, and session monitoring primitives for the documented Jules
API v1alpha resources.

## Install

```bash
go get github.com/SamyRai/go-jules@latest
```

## Use

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

	sessions, err := client.ListSessions(context.Background(), 30)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("found %d sessions", len(sessions))
}
```

## API Surface

- Sources: list and get connected source repositories.
- Sessions: create, list, get, delete, approve plans, and send messages.
- Activities: list, get, cursor helpers, and client-side filtering/search over
  documented activity payloads.
- Artifacts: read embedded `changeSet`, `bashOutput`, and base64 `media`
  payloads from activity responses.
- Monitoring: poll session state until completion, user action, or timeout.

The SDK intentionally avoids local filesystem side effects. Patch application,
artifact downloads, CLI presentation, and app-specific orchestration belong in
callers.

## Development

```bash
go mod tidy
go test ./...
go vet ./...
```

## License

MIT. See [LICENSE](LICENSE).
