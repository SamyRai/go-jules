# Release Notes and Process

`go-jules` is pre-1.0, but current consumers are treated as real. Breaking
changes should be explicit in release notes and reserved for API corrections or
architecture changes that materially improve maintainability.

Versioning follows Go module conventions:

- patch releases fix bugs without public API changes
- minor releases may add API surface or carry documented pre-1.0 breaking
  changes
- tagged versions are immutable; fixes are published as new tags

## Unreleased

- No unreleased changes.

## v0.2.0 - 2026-05-26

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
- Add scheduled API coverage validation in CI to catch Jules Discovery drift.
- Harden session monitoring with immediate first polls, positive duration
  validation, and explicit `WaitForPlan` activity-list errors.
- Add artifact kind and metadata helpers that classify embedded artifacts
  without decoding content or writing files.
- Add read-only live smoke coverage gated by `JULES_API_KEY`.

## Release Process

Releases are Go module releases. The package is published by pushing an
immutable semver git tag such as `v0.2.0`; consumers install it with:

```bash
go get github.com/SamyRai/go-jules@v0.2.0
```

Before tagging, verify from a clean main branch:

```bash
go mod tidy
go test ./...
go test -race ./...
go vet ./...
go run ./internal/cmd/jules-api-coverage
GOTOOLCHAIN=go1.26.3 go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Optionally run the read-only live smoke test with a maintainer API key:

```bash
JULES_API_KEY=... go test ./... -run TestLiveSmokeListEndpoints
```

Tag and publish:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

The release workflow validates the tag, repeats the verification suite, creates
or preserves the GitHub release, and checks that `proxy.golang.org` indexes the
tag.

## Packaging

This SDK does not publish to GitHub Packages and does not build release assets.
Go modules are resolved from git tags through the Go module proxy.

If a future release adds CLI binaries or archives, attach them to the GitHub
release with checksums:

```bash
gh release upload vX.Y.Z dist/* checksums.txt --clobber
```
