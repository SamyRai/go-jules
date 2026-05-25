package jules

import (
	"time"

	"github.com/SamyRai/go-jules/internal/resource"
	"github.com/SamyRai/go-jules/internal/services"
)

// NormalizeSessionName converts a bare session ID into a session resource name.
func NormalizeSessionName(session string) string {
	return resource.NormalizeSessionName(session)
}

// NormalizeSourceName converts a bare source ID into a source resource name.
func NormalizeSourceName(source string) string {
	return resource.NormalizeSourceName(source)
}

// ActivityCursor returns the latest createTime in the provided activities.
func ActivityCursor(activities []Activity) time.Time {
	return services.ActivityCursor(activities)
}

// FilterActivities applies client-side filters over documented activity fields.
func FilterActivities(activities []Activity, filter *ActivityFilter) []Activity {
	return services.FilterActivities(activities, filter)
}

// ArtifactContent returns the documented embedded content for an artifact.
func ArtifactContent(artifact Artifact) ([]byte, error) {
	return services.ArtifactContent(artifact)
}

// ArtifactKindOf returns the documented payload kind carried by an artifact.
func ArtifactKindOf(artifact Artifact) ArtifactKind {
	return services.ArtifactKindOf(artifact)
}

// ArtifactMetadataOf returns documented artifact metadata without decoding or
// writing artifact content.
func ArtifactMetadataOf(artifact Artifact) ArtifactMetadata {
	return services.ArtifactMetadataOf(artifact)
}

// NewSessionMonitor creates a new session monitor.
func NewSessionMonitor(client *Client, sessionID string) *SessionMonitor {
	return services.NewSessionMonitor(client.Sessions(), client.Activities(), sessionID)
}
