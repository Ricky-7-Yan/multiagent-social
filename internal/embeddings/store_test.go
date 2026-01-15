package embeddings

import (
	"context"
	"os"
	"testing"
)

func TestGenerateEmbedding_NoAPIKey(t *testing.T) {
	// ensure env not set
	os.Unsetenv("OPENAI_API_KEY")
	_, err := GenerateEmbedding(context.Background(), "hello world")
	if err == nil {
		t.Fatalf("expected error when OPENAI_API_KEY is not set")
	}
}

