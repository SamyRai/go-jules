// Package apicoverage validates this module's SDK surface against the Jules
// Discovery document.
//
// The package is maintainer tooling, not public SDK behavior. It loads a
// Discovery document, checks that documented operations are represented by SDK
// services, and verifies that mapped schema fields and enum values are present
// in the Go model types. The validation is intentionally explicit: newly
// documented Jules operations or schemas should fail the check until the SDK
// either supports them or records why they are operation-only response shells.
//
// Use [ValidateURL] to validate the live Discovery document and [ValidateReader]
// in tests or tools that already have Discovery JSON. The command in
// internal/cmd/jules-api-coverage is the canonical entry point for CI.
package apicoverage
