package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"strings"

	"github.com/SamyRai/go-jules/internal/model"
)

// ArtifactKind identifies the documented payload carried by an artifact.
type ArtifactKind string

const (
	ArtifactKindUnknown    ArtifactKind = ""
	ArtifactKindBashOutput ArtifactKind = "bashOutput"
	ArtifactKindChangeSet  ArtifactKind = "changeSet"
	ArtifactKindMedia      ArtifactKind = "media"
)

// ArtifactMetadata summarizes the documented metadata available without
// decoding or writing artifact content.
type ArtifactMetadata struct {
	Kind        ArtifactKind
	ContentType string
	Source      string
	Command     string
	ExitCode    int
	HasContent  bool
}

// ActivityArtifact represents an artifact with its activity context.
type ActivityArtifact struct {
	ActivityID string
	Index      int
	Artifact   model.Artifact
}

// ArtifactsService owns Jules artifact helpers.
type ArtifactsService struct {
	activities ActivityGetterLister
}

// ActivityGetterLister is the activity dependency artifacts need.
type ActivityGetterLister interface {
	Get(ctx context.Context, sessionID, activityID string) (*model.Activity, error)
	ListAll(ctx context.Context, sessionID string, pageSize int) ([]model.Activity, error)
}

func NewArtifactsService(activities ActivityGetterLister) *ArtifactsService {
	return &ArtifactsService{activities: activities}
}

// ListFromActivity retrieves all embedded artifacts from a documented
// activity response.
func (a *ArtifactsService) ListFromActivity(ctx context.Context, sessionID, activityID string) ([]model.Artifact, error) {
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
func (a *ArtifactsService) Content(artifact model.Artifact) ([]byte, error) {
	return artifactContent(artifact)
}

// Kind returns the documented artifact payload kind.
func (a *ArtifactsService) Kind(artifact model.Artifact) ArtifactKind {
	return ArtifactKindOf(artifact)
}

// Metadata returns documented artifact metadata without decoding or writing
// artifact content.
func (a *ArtifactsService) Metadata(artifact model.Artifact) ArtifactMetadata {
	return ArtifactMetadataOf(artifact)
}

// ArtifactContent returns the documented embedded content for an artifact.
func ArtifactContent(artifact model.Artifact) ([]byte, error) {
	return artifactContent(artifact)
}

// ArtifactKindOf returns the documented payload kind carried by an artifact.
func ArtifactKindOf(artifact model.Artifact) ArtifactKind {
	switch {
	case artifact.BashOutput != nil:
		return ArtifactKindBashOutput
	case artifact.ChangeSet != nil:
		return ArtifactKindChangeSet
	case artifact.Media != nil:
		return ArtifactKindMedia
	default:
		return ArtifactKindUnknown
	}
}

// ArtifactMetadataOf returns documented artifact metadata without decoding or
// writing artifact content.
func ArtifactMetadataOf(artifact model.Artifact) ArtifactMetadata {
	metadata := ArtifactMetadata{Kind: ArtifactKindOf(artifact)}
	switch {
	case artifact.BashOutput != nil:
		metadata.ContentType = "text/plain; charset=utf-8"
		metadata.Command = artifact.BashOutput.Command
		metadata.ExitCode = artifact.BashOutput.ExitCode
		metadata.HasContent = artifact.BashOutput.Command != "" || artifact.BashOutput.Output != ""
	case artifact.ChangeSet != nil:
		metadata.ContentType = "text/x-diff; charset=utf-8"
		metadata.Source = artifact.ChangeSet.Source
		metadata.HasContent = artifact.ChangeSet.GitPatch != nil && artifact.ChangeSet.GitPatch.UnidiffPatch != ""
	case artifact.Media != nil:
		metadata.ContentType = artifact.Media.MimeType
		if metadata.ContentType != "" {
			if parsed, _, err := mime.ParseMediaType(metadata.ContentType); err == nil {
				metadata.ContentType = parsed
			}
		}
		metadata.HasContent = artifact.Media.Data != ""
	}
	return metadata
}

func artifactContent(artifact model.Artifact) ([]byte, error) {
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
