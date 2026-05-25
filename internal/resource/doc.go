// Package resource owns Jules resource-name normalization and path escaping.
//
// Services accept both bare IDs and full resource names at the public facade.
// This package centralizes the conversion to canonical names and escaped URL
// paths so every service applies the same validation rules:
//
//   - sessions are single-segment resources under "sessions/"
//   - sources can contain multiple non-empty path segments under "sources/"
//   - activities belong to a session and can be addressed by bare ID or full
//     "sessions/{session}/activities/{activity}" name
//
// The package does not build complete URLs and does not know the API base URL.
// Service packages append the returned escaped paths to their configured base
// URL.
package resource
