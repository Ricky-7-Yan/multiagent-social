package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/yourname/multiagent-social/internal/agent"
	"github.com/yourname/multiagent-social/internal/embeddings"
	"github.com/yourname/multiagent-social/internal/persistence"
	"github.com/yourname/multiagent-social/internal/pubsub"
)

// Orchestrator coordinates conversations and agent actions.
type Orchestrator struct {
	store         *persistence.PostgresStore
	ps            *pubsub.RedisPubSub
	responseDelay time.Duration
}

// NewOrchestrator constructs an orchestrator instance.
func NewOrchestrator(store *persistence.PostgresStore, ps *pubsub.RedisPubSub) *Orchestrator {
	return &Orchestrator{
		store:         store,
		ps:            ps,
		responseDelay: 500 * time.Millisecond,
	}
}

// CreateConversation creates a conversation row and returns its id.
func (o *Orchestrator) CreateConversation(ctx context.Context, title string, agentIDs []string) (string, error) {
	id, err := o.store.CreateConversation(ctx, title)
	if err != nil {
		return "", err
	}
	// publish event for consumers
	_ = o.ps.Publish(ctx, fmt.Sprintf("conversation:%s", id), map[string]interface{}{
		"event": "conversation.created",
		"id":    id,
		"title": title,
	})
	return id, nil
}

// HandleUserMessage stores the user message and schedules agent responses in turn order.
func (o *Orchestrator) HandleUserMessage(ctx context.Context, conversationID string, userID string, content string) error {
	// persist the user message
	msgID, err := o.store.InsertMessage(ctx, conversationID, "user", userID, content)
	if err != nil {
		return err
	}
	// publish user message event
	_ = o.ps.Publish(ctx, fmt.Sprintf("conversation:%s", conversationID), map[string]interface{}{
		"event":   "message.created",
		"sender":  userID,
		"content": content,
	})
	// generate and save embedding for this message asynchronously using OpenAI
	go func() {
		if vec, err := embeddings.GenerateEmbedding(context.Background(), content); err == nil {
			_ = embeddings.SaveEmbedding(context.Background(), o.store, conversationID, msgID, vec)
		}
	}()

	// run agent responses asynchronously so request returns fast
	go o.scheduleAgentResponses(context.Background(), conversationID)
	return nil
}

// scheduleAgentResponses loads agents, gets context, and runs responses in sequence (turns).
func (o *Orchestrator) scheduleAgentResponses(ctx context.Context, conversationID string) {
	agents, err := o.store.ListAgents(ctx)
	if err != nil || len(agents) == 0 {
		return
	}
	// gather current messages as context
	messages, err := o.store.GetConversationMessages(ctx, conversationID)
	if err != nil {
		return
	}
	// choose up to 3 agents in order (simple selection: first N)
	limit := 3
	for i, a := range agents {
		if i >= limit {
			break
		}
		decider := &agent.SimpleDecider{}
		action, derr := decider.DecideAction(ctx, &a, &agent.ConversationState{
			ConversationID: conversationID,
			Messages:       messages,
		})
		if derr != nil || action == nil {
			continue
		}
		// persist agent message
		msgID, merr := o.store.InsertMessage(ctx, conversationID, "agent", string(a.ID), action.Payload)
		if merr != nil {
			continue
		}
		// save embedding for agent message using OpenAI
		go func(mid string, payload string) {
			if vec, err := embeddings.GenerateEmbedding(context.Background(), payload); err == nil {
				_ = embeddings.SaveEmbedding(context.Background(), o.store, conversationID, mid, vec)
			}
		}(msgID, action.Payload)
		// publish agent speak event
		_ = o.ps.Publish(ctx, fmt.Sprintf("conversation:%s", conversationID), map[string]interface{}{
			"event":   "message.created",
			"sender":  a.Name,
			"content": action.Payload,
		})
		// append to messages for next agent context
		messages = append(messages, action.Payload)
		// wait a bit to simulate turn-taking
		time.Sleep(o.responseDelay)
	}
}

// StartDebate starts a structured debate between selected agents for given rounds.
func (o *Orchestrator) StartDebate(ctx context.Context, conversationID string, participantIDs []string, rounds int) error {
	if rounds <= 0 {
		rounds = 3
	}
	// load agents
	agents, err := o.store.ListAgents(ctx)
	if err != nil {
		return err
	}
	// map agents by id
	agentMap := make(map[string]agent.Agent)
	for _, a := range agents {
		agentMap[string(a.ID)] = a
	}
	// filter participants
	var participants []agent.Agent
	for _, pid := range participantIDs {
		if a, ok := agentMap[pid]; ok {
			participants = append(participants, a)
		}
	}
	if len(participants) < 2 {
		return fmt.Errorf("need at least two participants")
	}

	// initial context
	messages, err := o.store.GetConversationMessages(ctx, conversationID)
	if err != nil {
		return err
	}
	topic := "讨论"
	if len(messages) > 0 {
		topic = messages[len(messages)-1]
	}

	// run debate rounds: each participant speaks in order per round
	for r := 0; r < rounds; r++ {
		for _, p := range participants {
			// generate a debate-style payload
			payload := fmt.Sprintf("%s（第%d轮）: 我对%s的看法是基于我的身份[%s]，我认为...", p.Name, r+1, topic, p.Persona)
			// persist
			msgID, merr := o.store.InsertMessage(ctx, conversationID, "agent", string(p.ID), payload)
			if merr != nil {
				continue
			}
			_ = o.ps.Publish(ctx, fmt.Sprintf("conversation:%s", conversationID), map[string]interface{}{
				"event":   "message.created",
				"sender":  p.Name,
				"content": payload,
			})
			// save embedding
			go func(mid, content string) {
				if vec, err := embeddings.GenerateEmbedding(context.Background(), content); err == nil {
					_ = embeddings.SaveEmbedding(context.Background(), o.store, conversationID, mid, vec)
				}
			}(msgID, payload)
			time.Sleep(o.responseDelay)
		}
	}
	return nil
}

