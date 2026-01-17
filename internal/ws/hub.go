package ws

import (
	"context"
	"encoding/json"
	"log"
)

// MessageEvent is a generic event structure used over pubsub.
type MessageEvent struct {
	Event   string `json:"event"`
	Sender  string `json:"sender,omitempty"`
	Content string `json:"content,omitempty"`
}

// StartSubscriber listens on Redis pubsub channel and sends to provided handler.
// Subscriber is a minimal interface for pubsub subscribers used by ws package.
type Subscriber interface {
	Channel() <-chan string
	Close() error
}

// StartSubscriber listens on a generic Subscriber (string payloads) and sends parsed events to handler.
func StartSubscriber(ctx context.Context, sub Subscriber, handler func(MessageEvent)) {
	go func() {
		defer sub.Close()
		ch := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case payload, ok := <-ch:
				if !ok {
					return
				}
				var evt MessageEvent
				if err := json.Unmarshal([]byte(payload), &evt); err != nil {
					log.Printf("ws: failed to unmarshal event: %v", err)
					continue
				}
				handler(evt)
			}
		}
	}()
}

