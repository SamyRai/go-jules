package jules_test

import (
	"testing"

	jules "github.com/SamyRai/go-jules"
)

func TestPublicFacadeUsesSingleImportPath(t *testing.T) {
	client := jules.NewClient("api-key")

	if client.Config().APIKey != "api-key" {
		t.Fatalf("Config().APIKey = %q, want api-key", client.Config().APIKey)
	}

	var _ *jules.SessionsService = client.Sessions()
	var _ *jules.SourcesService = client.Sources()
	var _ *jules.ActivitiesService = client.Activities()
	var _ *jules.ArtifactsService = client.Artifacts()
}
