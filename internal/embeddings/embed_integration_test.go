package embeddings

import (
	"context"
	"os"
	"testing"

	"github.com/yourname/multiagent-social/internal/persistence"
)

// This integration test requires a running Postgres with pgvector and PG_DSN env set.
func TestEmbeddingsIntegration(t *testing.T) {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		t.Skip("PG_DSN not set; skipping integration test")
	}
	store, err := persistence.NewPostgresStore(context.Background(), dsn)
	if err != nil {
		t.Skipf("db not available: %v", err)
	}
	defer store.Close()

	// create conversation and message
	convID, err := store.CreateConversation(context.Background(), "it-test")
	if err != nil {
		t.Fatalf("create conv: %v", err)
	}
	msgID, err := store.InsertMessage(context.Background(), convID, "user", "it", "hello")
	if err != nil {
		t.Fatalf("insert msg: %v", err)
	}
	vec := Vector{0.1, 0.2, 0.3, 0.4}
	if err := SaveEmbedding(context.Background(), store, convID, msgID, vec); err != nil {
		t.Fatalf("save embed: %v", err)
	}
	ids, err := QuerySimilarMessages(context.Background(), store, vec, 5)
	if err != nil {
		t.Fatalf("query similar: %v", err)
	}
	if len(ids) == 0 {
		t.Fatalf("expected >=1 similar id")
	}
}

