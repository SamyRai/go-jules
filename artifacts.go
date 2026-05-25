package jules

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
)

// ActivityArtifact represents an artifact with its activity context.
type ActivityArtifact struct {
	ActivityID string
	Index      int
	Artifact   Artifact
}

// ArtifactsService owns Jules artifact helpers.
type ArtifactsService struct {
	activities *ActivitiesService
}

// ListFromActivity retrieves all embedded artifacts from a documented
// activity response.
func (a *ArtifactsService) ListFromActivity(ctx context.Context, sessionID, activityID string) ([]Artifact, error) {
	activity, err := a.activities.Get(ctx, sessionID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return activity.Artifacts, nil
}

// ListFromSession retrieves all embedded artifacts from all documented
// activity responses in a session.
func (a *ArtifactsService) ListFromSession(ctx context.Context, sessionID string) ([]ActivityArtifact, error) {
	activities, err := a.activities.ListAll(ctx, sessionID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	var allArtifacts []ActivityArtifact
	for _, activity := range activities {
		for i, artifact := range activity.Artifacts {
			allArtifacts = append(allArtifacts, ActivityArtifact{
				ActivityID: activity.ID,
				Index:      i,
				Artifact:   artifact,
			})
		}
	}

	return allArtifacts, nil
}

// Content returns the documented embedded content for an artifact.
func (a *ArtifactsService) Content(artifact Artifact) ([]byte, error) {
	return artifactContent(artifact)
}

// ArtifactContent returns the documented embedded content for an artifact.
func ArtifactContent(artifact Artifact) ([]byte, error) {
	return artifactContent(artifact)
}

func artifactContent(artifact Artifact) ([]byte, error) {
	switch {
	case artifact.BashOutput != nil:
		var builder strings.Builder
		if artifact.BashOutput.Command != "" {
			builder.WriteString("$ ")
			builder.WriteString(artifact.BashOutput.Command)
			builder.WriteString("\n")
		}
		if artifact.BashOutput.Output != "" {
			builder.WriteString(artifact.BashOutput.Output)
			if !strings.HasSuffix(artifact.BashOutput.Output, "\n") {
				builder.WriteString("\n")
			}
		}
		builder.WriteString(fmt.Sprintf("exit code: %d\n", artifact.BashOutput.ExitCode))
		return []byte(builder.String()), nil
	case artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil:
		return []byte(artifact.ChangeSet.GitPatch.UnidiffPatch), nil
	case artifact.Media != nil:
		data, err := base64.StdEncoding.DecodeString(artifact.Media.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode media artifact: %w", err)
		}
		return data, nil
	default:
		return nil, fmt.Errorf("artifact has no documented content")
	}
}
