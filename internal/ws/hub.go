package ws

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

// MessageEvent is a generic event structure used over pubsub.
type MessageEvent struct {
	Event   string `json:"event"`
	Sender  string `json:"sender,omitempty"`
	Content string `json:"content,omitempty"`
}

// StartSubscriber listens on Redis pubsub channel and sends to provided handler.
func StartSubscriber(ctx context.Context, client *redis.Client, channel string, handler func(MessageEvent)) {
	sub := client.Subscribe(ctx, channel)
	go func() {
		defer sub.Close()
		ch := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var evt MessageEvent
				if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
					log.Printf("ws: failed to unmarshal event: %v", err)
					continue
				}
				handler(evt)
			}
		}
	}()
}

