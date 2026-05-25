package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourcePaths(t *testing.T) {
	sessionPath, err := SessionPath("sessions/session 1")
	require.NoError(t, err)
	assert.Equal(t, "sessions/session%201", sessionPath)

	sourcePath, err := SourcePath("sources/github/owner name/repo name")
	require.NoError(t, err)
	assert.Equal(t, "sources/github/owner%20name/repo%20name", sourcePath)

	activityPath, err := ActivityPath("ignored", "sessions/session 1/activities/activity 1")
	require.NoError(t, err)
	assert.Equal(t, "sessions/session%201/activities/activity%201", activityPath)
}

func TestResourcePathValidation(t *testing.T) {
	_, err := SessionPath("sessions/a/b")
	require.Error(t, err)

	_, err = SourcePath("sources/github//repo")
	require.Error(t, err)

	_, err = ActivityPath("sessions/123", "activities/a/b")
	require.Error(t, err)
}
