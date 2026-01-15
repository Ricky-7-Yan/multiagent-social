package agent

import (
	"context"
	"errors"
	"time"
)

// AgentID is a simple string alias for agent identity
type AgentID string

// Agent represents a virtual persona participating in conversations.
type Agent struct {
	ID              AgentID
	Name            string
	Persona         string                 // textual persona description
	BehaviorProfile map[string]interface{} // tunable behavior knobs
	CreatedAt       time.Time
}

// ConversationState is a lightweight snapshot of the conversation for decision making.
type ConversationState struct {
	ConversationID string
	Messages       []string
}

// Action represents what an Agent wants to do.
type Action struct {
	Type    string // "speak", "ask", "challenge"
	Payload string
}

// Decider returns an action for an agent given conversation state.
type Decider interface {
	DecideAction(ctx context.Context, a *Agent, state *ConversationState) (*Action, error)
}

// SimpleDecider is an example decider that echoes last message or introduces a topic.
type SimpleDecider struct{}

func (s *SimpleDecider) DecideAction(ctx context.Context, a *Agent, state *ConversationState) (*Action, error) {
	if a == nil {
		return nil, errors.New("nil agent")
	}
	// If there are messages, reply by reflecting last one; otherwise introduce self.
	if len(state.Messages) > 0 {
		last := state.Messages[len(state.Messages)-1]
		return &Action{
			Type:    "speak",
			Payload: a.Name + "回应: " + last,
		}, nil
	}
	return &Action{
		Type:    "speak",
		Payload: a.Name + " 你好，我可以谈论 " + a.Persona,
	}, nil
}

