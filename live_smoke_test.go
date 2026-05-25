package jules

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestLiveSmokeListEndpoints(t *testing.T) {
	apiKey := os.Getenv("JULES_API_KEY")
	if apiKey == "" {
		t.Skip("set JULES_API_KEY to run live Jules API smoke tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewClient(apiKey, WithRetryAttempts(1))

	if _, err := client.Sources().List(ctx, &ListSourcesOptions{PageSize: 1}); err != nil {
		t.Fatalf("list sources: %v", err)
	}
	if _, err := client.Sessions().List(ctx, &ListSessionsOptions{PageSize: 1}); err != nil {
		t.Fatalf("list sessions: %v", err)
	}
}
