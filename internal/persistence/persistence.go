package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourname/multiagent-social/internal/agent"
)

// PostgresStore is a thin wrapper around pgxpool for this project.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore creates and verifies the DB connection.
func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	// quick ping
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// Pool exposes the underlying pgxpool for advanced operations.
func (s *PostgresStore) Pool() *pgxpool.Pool {
	return s.pool
}

// ListAgents returns all agents for MVP (lightweight).
func (s *PostgresStore) ListAgents(ctx context.Context) ([]agent.Agent, error) {
	rows, err := s.pool.Query(ctx, "SELECT id, name, persona, behavior_profile, created_at FROM agents")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []agent.Agent
	for rows.Next() {
		var id string
		var name string
		var personaBytes []byte
		var behaviorBytes []byte
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &personaBytes, &behaviorBytes, &createdAt); err != nil {
			return nil, err
		}
		persona := string(personaBytes)
		var behaviorProfile map[string]interface{}
		if len(behaviorBytes) > 0 {
			_ = json.Unmarshal(behaviorBytes, &behaviorProfile)
		}
		out = append(out, agent.Agent{
			ID:              agent.AgentID(id),
			Name:            name,
			Persona:         persona,
			BehaviorProfile: behaviorProfile,
			CreatedAt:       createdAt,
		})
	}
	return out, nil
}

// CreateAgent inserts a new agent row and returns its id.
func (s *PostgresStore) CreateAgent(ctx context.Context, name string, persona string, behaviorProfile map[string]interface{}) (string, error) {
	var id string
	var behaviorBytes []byte
	if behaviorProfile != nil {
		var err error
		behaviorBytes, err = json.Marshal(behaviorProfile)
		if err != nil {
			return "", err
		}
	}
	err := s.pool.QueryRow(ctx, "INSERT INTO agents (name, persona, behavior_profile) VALUES ($1, $2, $3) RETURNING id", name, persona, behaviorBytes).Scan(&id)
	return id, err
}

// CreateConversation creates an empty conversation row and returns id.
func (s *PostgresStore) CreateConversation(ctx context.Context, title string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, "INSERT INTO conversations (title) VALUES ($1) RETURNING id", title).Scan(&id)
	return id, err
}

// InsertMessage persists a message to messages table.
func (s *PostgresStore) InsertMessage(ctx context.Context, conversationID, senderType, senderID, content string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, "INSERT INTO messages (conversation_id, sender_type, sender_id, content) VALUES ($1, $2, $3, $4) RETURNING id", conversationID, senderType, senderID, content).Scan(&id)
	return id, err
}

// GetConversationMessages returns messages content for a conversation.
func (s *PostgresStore) GetConversationMessages(ctx context.Context, conversationID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, "SELECT content FROM messages WHERE conversation_id=$1 ORDER BY created_at ASC", conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			return nil, err
		}
		out = append(out, content)
	}
	return out, nil
}

// ListConversations returns id and title for recent conversations.
func (s *PostgresStore) ListConversations(ctx context.Context) ([]struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}, error) {
	rows, err := s.pool.Query(ctx, "SELECT id, title FROM conversations ORDER BY created_at DESC LIMIT 100")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	for rows.Next() {
		var id string
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			return nil, err
		}
		out = append(out, struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		}{ID: id, Title: title})
	}
	return out, nil
}

