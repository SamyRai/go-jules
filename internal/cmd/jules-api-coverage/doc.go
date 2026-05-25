// Jules-api-coverage validates this module's Jules API coverage.
//
// The command fetches the Jules v1alpha Discovery document, compares it with
// the SDK's supported operations and mapped schemas, and exits non-zero when
// coverage drifts. It is intended for maintainers and CI:
//
//	go run ./internal/cmd/jules-api-coverage
//
// Importing applications should not depend on this command or on the internal
// validation packages it uses.
package main
