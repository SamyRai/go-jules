// Package services owns Jules API resource services and higher-level helpers.
//
// Services are the SDK's application-layer boundary. Each service owns one
// Jules resource area and depends only on a JSON transport interface plus the
// API base URL:
//
//   - SourcesService lists and retrieves connected source repositories.
//   - SessionsService creates, lists, retrieves, archives, unarchives, deletes,
//     and messages sessions.
//   - ActivitiesService lists, retrieves, filters, and searches documented
//     session activity payloads.
//   - ArtifactsService extracts documented embedded artifact content from
//     activities.
//   - SessionMonitor polls session and activity state through narrow service
//     interfaces rather than through the public client facade.
//
// Cross-service dependencies are deliberately narrow. For example, session
// creation uses a SourceResolver only to infer a missing starting branch for
// source-backed sessions, and artifact helpers depend only on activity get/list
// behavior. Public consumers access these implementations through the root
// jules package.
package services
