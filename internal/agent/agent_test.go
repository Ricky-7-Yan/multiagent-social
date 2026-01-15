package agent

import (
	"context"
	"testing"
)

func TestSimpleDeciderReplies(t *testing.T) {
	dec := &SimpleDecider{}
	a := &Agent{ID: "a1", Name: "TestAgent", Persona: "music, literature"}
	state := &ConversationState{
		ConversationID: "conv1",
		Messages:       []string{"hello"},
	}
	act, err := dec.DecideAction(context.Background(), a, state)
	if err != nil {
		t.Fatal(err)
	}
	if act == nil || act.Payload == "" {
		t.Fatalf("expected payload, got %+v", act)
	}
}

