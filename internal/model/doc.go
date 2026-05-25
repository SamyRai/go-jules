// Package model owns Jules API data transfer objects, enum values, and API error
// types.
//
// The structs in this package mirror documented Jules JSON payloads. They are
// intentionally behavior-light: resource operations, transport retries,
// normalization, and monitoring live in other internal packages. Keeping models
// free of HTTP and service concerns makes schema coverage validation and public
// facade aliases straightforward.
//
// Public consumers receive these types through aliases in the root jules package
// instead of importing this internal package directly. When adding fields, keep
// JSON tags aligned with the Discovery document and update the coverage mapping
// in internal/apicoverage when the new schema is externally documented.
package model
