package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yourname/multiagent-social/internal/agent"
)

type InMemoryStore struct {
	mu            sync.RWMutex
	agents        map[string]agent.Agent
	conversations map[string]string   // id -> title
	messages      map[string][]string // convID -> messages
	counter       int64
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		agents:        make(map[string]agent.Agent),
		conversations: make(map[string]string),
		messages:      make(map[string][]string),
		counter:       time.Now().UnixNano(),
	}
}

func (s *InMemoryStore) nextID(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	return fmt.Sprintf("%s-%d", prefix, s.counter)
}

func (s *InMemoryStore) ListAgents() []agent.Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]agent.Agent, 0, len(s.agents))
	for _, a := range s.agents {
		out = append(out, a)
	}
	return out
}

func (s *InMemoryStore) CreateAgent(name, persona string) string {
	id := s.nextID("agent")
	a := agent.Agent{
		ID:      agent.AgentID(id),
		Name:    name,
		Persona: persona,
	}
	s.mu.Lock()
	s.agents[id] = a
	s.mu.Unlock()
	return id
}

func (s *InMemoryStore) GetAgent(id string) (agent.Agent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.agents[id]
	return a, ok
}

func (s *InMemoryStore) UpdateAgent(id, name, persona string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.agents[id]
	if !ok {
		return false
	}
	a.Name = name
	a.Persona = persona
	s.agents[id] = a
	return true
}

func (s *InMemoryStore) DeleteAgent(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.agents[id]
	if !ok {
		return false
	}
	delete(s.agents, id)
	return true
}

func (s *InMemoryStore) CreateConversation(title string) string {
	id := s.nextID("conv")
	s.mu.Lock()
	s.conversations[id] = title
	s.messages[id] = []string{}
	s.mu.Unlock()
	return id
}

func (s *InMemoryStore) InsertMessage(convID, sender, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[convID] = append(s.messages[convID], fmt.Sprintf("%s: %s", sender, content))
}

func (s *InMemoryStore) GetConversationMessages(convID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cpy := make([]string, len(s.messages[convID]))
	copy(cpy, s.messages[convID])
	return cpy
}

// SSE subscribers for devserver (conversation events)
var (
	subsMu      sync.Mutex
	subscribers = map[string][]chan string{}
)

func subscribeEvents(convID string) <-chan string {
	ch := make(chan string, 8)
	subsMu.Lock()
	subscribers[convID] = append(subscribers[convID], ch)
	subsMu.Unlock()
	return ch
}

func publishEvent(convID string, msg string) {
	subsMu.Lock()
	defer subsMu.Unlock()
	for _, ch := range subscribers[convID] {
		select {
		case ch <- msg:
		default:
		}
	}
}

func main() {
	store := NewInMemoryStore()
	// seed one agent
	store.CreateAgent("Alice", "music, literature")
	store.CreateAgent("Bob", "fitness, philosophy")

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/api/v1/agents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			agents := store.ListAgents()
			// normalize to lowercase keys for the simple admin UI
			out := make([]map[string]interface{}, 0, len(agents))
			for _, a := range agents {
				out = append(out, map[string]interface{}{
					"id":               a.ID,
					"name":             a.Name,
					"persona":          a.Persona,
					"behavior_profile": a.BehaviorProfile,
					"created_at":       a.CreatedAt,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(out)
			return
		}
		if r.Method == http.MethodPost {
			var p struct {
				Name    string `json:"name"`
				Persona string `json:"persona"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}
			id := store.CreateAgent(p.Name, p.Persona)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]string{"id": id})
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	// GET/PUT/DELETE for single agent: /api/v1/agents/{id}
	mux.HandleFunc("/api/v1/agents/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			a, ok := store.GetAgent(id)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			out := map[string]interface{}{"id": a.ID, "name": a.Name, "persona": a.Persona, "behavior_profile": a.BehaviorProfile}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(out)
			return
		case http.MethodPut:
			var p struct {
				Name    string `json:"name"`
				Persona string `json:"persona"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}
			if ok := store.UpdateAgent(id, p.Name, p.Persona); !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodDelete:
			if ok := store.DeleteAgent(id); !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})

	mux.HandleFunc("/api/v1/conversations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			id := store.CreateConversation("Conversation (dev)")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(id))
			return
		}
		if r.Method == http.MethodGet {
			// list convs
			store.mu.RLock()
			out := make([]map[string]string, 0, len(store.conversations))
			for id, title := range store.conversations {
				out = append(out, map[string]string{"id": id, "title": title})
			}
			store.mu.RUnlock()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(out)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/api/v1/conversations/", func(w http.ResponseWriter, r *http.Request) {
		// expecting /api/v1/conversations/{id}/messages
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/conversations/"), "/")
		if len(parts) < 2 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		convID := parts[0]
		action := parts[1]
		if action != "messages" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method == http.MethodPost {
			defer r.Body.Close()
			b, _ := io.ReadAll(r.Body)
			content := string(b)
			if content == "" {
				http.Error(w, "empty body", http.StatusBadRequest)
				return
			}
			store.InsertMessage(convID, "user", content)
			// publish SSE event for subscribers
			publishEvent(convID, fmt.Sprintf(`{"sender":"user","content":%q}`, content))
			// simple dev orchestrator: pick up to 2 agents and reply
			agents := store.ListAgents()
			limit := 2
			if len(agents) < limit {
				limit = len(agents)
			}
			for i := 0; i < limit; i++ {
				a := agents[i]
				dec := &agent.SimpleDecider{}
				act, err := dec.DecideAction(r.Context(), &a, &agent.ConversationState{
					ConversationID: convID,
					Messages:       store.GetConversationMessages(convID),
				})
				if err != nil || act == nil {
					continue
				}
				store.InsertMessage(convID, a.Name, act.Payload)
				// publish SSE event for agent reply
				publishEvent(convID, fmt.Sprintf(`{"sender":%q,"content":%q}`, a.Name, act.Payload))
			}
			w.WriteHeader(http.StatusAccepted)
			return
		}
		if r.Method == http.MethodGet {
			msgs := store.GetConversationMessages(convID)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(msgs)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	// serve admin static files from web/admin
	adminDir := filepath.Join("web", "admin")
	mux.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(http.Dir(adminDir))))

	// SSE endpoint for conversation events (dev)
	mux.HandleFunc("/events/conversations/", func(w http.ResponseWriter, r *http.Request) {
		// expecting /events/conversations/{id}
		id := strings.TrimPrefix(r.URL.Path, "/events/conversations/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		ch := subscribeEvents(id)
		notify := r.Context().Done()
		for {
			select {
			case <-notify:
				return
			case m := <-ch:
				_, _ = fmt.Fprintf(w, "data: %s\n\n", m)
				flusher.Flush()
			}
		}
	})

	// wrap with simple CORS middleware
	handler := allowCORS(mux)

	addr := ":8080"
	log.Printf("dev server starting on %s (DEV_MODE)", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// allowCORS sets permissive CORS headers for development.
func allowCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

