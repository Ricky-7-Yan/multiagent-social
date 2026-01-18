package ws

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"net/http"
	"os"
	"time"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/yourname/multiagent-social/internal/orchestrator"
	"github.com/yourname/multiagent-social/internal/persistence"
	"github.com/yourname/multiagent-social/internal/pubsub"
)

// HandleConversationWS returns an HTTP handler that upgrades to WebSocket
// and subscribes to Redis pubsub for conversation events. It also replays recent history.
func HandleConversationWS(orch *orchestrator.Orchestrator, ps *pubsub.RedisPubSub, store *persistence.PostgresStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Expect path: /ws/conversations/{id}
		convID := strings.TrimPrefix(r.URL.Path, "/ws/conversations/")
		convID = strings.Trim(convID, "/")
		if convID == "" {
			http.Error(w, "missing conversation id", http.StatusBadRequest)
			return
		}
		// simple token auth (optional)
		token := r.URL.Query().Get("token")
		if expected := r.Header.Get("X-WS-Token"); expected != "" {
			// header takes precedence
			token = expected
		}
		if required := os.Getenv("AUTH_TOKEN"); required != "" {
			if token == "" || token != required {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}

		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			Subprotocols: []string{},
		})
		if err != nil {
			http.Error(w, "failed to upgrade websocket", http.StatusInternalServerError)
			return
		}
		defer c.Close(websocket.StatusNormalClosure, "")

		// subscribe to redis events for this conversation
		redisClient := ps.Client()
		ch := redisClient.Subscribe(r.Context(), "conversation:"+convID)
		defer ch.Close()

		// send recent history to the client
		if store != nil {
			if msgs, err := store.GetConversationMessages(r.Context(), convID); err == nil {
				for _, m := range msgs {
					_ = wsjson.Write(r.Context(), c, map[string]interface{}{
						"event":   "history",
						"content": m,
					})
				}
			}
		}

		// read incoming (optional)
		go func() {
			for {
				var v interface{}
				if err := wsjson.Read(r.Context(), c, &v); err != nil {
					return
				}
				// ignore incoming messages in this MVP
				_ = v
			}
		}()

		// heartbeat: send ping events periodically
		pingCtx, pingCancel := context.WithCancel(r.Context())
		go func() {
			t := time.NewTicker(30 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-pingCtx.Done():
					return
				case <-t.C:
					_ = wsjson.Write(r.Context(), c, map[string]interface{}{
						"event": "ping",
						"ts":    time.Now().UTC().String(),
					})
				}
			}
		}()
		defer pingCancel()

		// forward redis messages to websocket
		for {
			msg, ok := <-ch.Channel()
			if !ok {
				return
			}
			var evt map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
				log.Printf("ws: unmarshal err: %v", err)
				continue
			}
			if err := wsjson.Write(r.Context(), c, evt); err != nil {
				return
			}
		}
	}
}

