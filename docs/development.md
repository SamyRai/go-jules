# Development Guide

`go-jules` is a Go module for the Jules REST API. It requires Go 1.25 or newer.
CI tests Go 1.25.x and Go 1.26.x, and release verification uses Go 1.26.3.

## Repository Layout

- `doc.go`, `client.go`, `types.go`, and `helpers.go` define the public
  `github.com/SamyRai/go-jules` package.
- `internal/model` owns Jules DTOs and `APIError`.
- `internal/resource` owns resource-name normalization and request path
  escaping.
- `internal/transport` owns HTTP request execution, retries, API errors, and
  debug logging.
- `internal/services` owns Sources, Sessions, Activities, Artifacts, pagination,
  and monitoring behavior.
- `internal/apicoverage` validates SDK coverage against the Jules Discovery
  document.
- `internal/cmd/jules-api-coverage` is the maintainer command used by CI and
  releases.
- `testdata` contains JSON fixtures used by tests.

Keep the public API in the root package. Internal packages should have one
clear owner of behavior and should not be imported by consumers.

## Local Verification

Run the same checks CI expects:

```bash
gofmt -s -l .
go mod tidy
go test ./...
go test -race ./...
go vet ./...
go run ./internal/cmd/jules-api-coverage
GOTOOLCHAIN=go1.26.3 go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

`go mod tidy` should leave `go.mod` and `go.sum` unchanged. If formatting output
is non-empty, run `gofmt -s -w` on the listed files before retesting.

Live smoke tests are opt-in and read-only. To run them, set `JULES_API_KEY` and
run:

```bash
JULES_API_KEY=... go test ./... -run TestLiveSmokeListEndpoints
```

The smoke test lists one source page and one session page. It does not create,
delete, archive, approve, message, write artifacts, or apply patches.

## API Coverage

The API coverage command fetches the Jules v1alpha Discovery document from:

```text
https://jules.googleapis.com/$discovery/rest?version=v1alpha
```

It validates:

- documented operations against SDK service methods
- documented paths, HTTP methods, and parameters
- documented schemas against exported public DTO aliases
- selected enum values against SDK constants

When the Discovery document changes, update the implementation and
`internal/apicoverage/coverage.go` together. Do not mark a newly documented API
as covered unless the corresponding public service method or DTO exists.

## CI

The CI workflow runs on pushes and pull requests to `main`, manual dispatch, and
a weekly schedule that catches Jules API drift even when repository code has not
changed. It checks formatting, module tidiness, tests, race tests, vet, API
coverage on Go 1.26.x, and govulncheck on Go 1.26.3.

The release workflow runs on semver tags and manual dispatch. It repeats the
release verification suite, validates the module path, creates a GitHub release
when needed, and checks Go proxy indexing.

## Documentation

Keep root documentation limited to `README.md`, `SECURITY.md`, `LICENSE`, and
other core policy files. Put additional guides under `docs/` and link them from
`docs/README.md` and the root README.
