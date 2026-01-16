package embeddings

import (
	"context"
	"errors"
	"os"

	openai "github.com/sashabaranov/go-openai"
	"github.com/pgvector/pgvector-go"

	"github.com/yourname/multiagent-social/internal/persistence"
)

// Vector is a slice of float32 representing an embedding.
type Vector = []float32

// GenerateEmbedding calls OpenAI embeddings API and returns a vector.
func GenerateEmbedding(ctx context.Context, text string) (Vector, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil, errors.New("OPENAI_API_KEY not set")
	}
	client := openai.NewClient(key)
	resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.AdaEmbeddingV2,
		Input: []string{text},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, errors.New("no embedding returned")
	}
	vec := make(Vector, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		vec[i] = float32(v)
	}
	return vec, nil
}

// SaveEmbedding stores the vector into embeddings table using pgvector.
func SaveEmbedding(ctx context.Context, store *persistence.PostgresStore, conversationID, messageID string, v Vector) error {
	if store == nil {
		return errors.New("nil store")
	}
	pgvec := pgvector.NewVector(v)
	_, err := store.Pool().Exec(ctx, "INSERT INTO embeddings (conversation_id, message_id, vector) VALUES ($1, $2, $3)", conversationID, messageID, pgvec)
	return err
}

// QuerySimilarMessages returns up to k message IDs similar to the provided vector using pgvector distance operator.
func QuerySimilarMessages(ctx context.Context, store *persistence.PostgresStore, vector Vector, k int) ([]string, error) {
	if store == nil {
		return nil, errors.New("nil store")
	}
	pgvec := pgvector.NewVector(vector)
	// use pgvector '<->' distance operator (lower = more similar)
	rows, err := store.Pool().Query(ctx, "SELECT message_id FROM embeddings ORDER BY vector <-> $1 LIMIT $2", pgvec, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

