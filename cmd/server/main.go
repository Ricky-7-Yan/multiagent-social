package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/yourname/multiagent-social/internal/orchestrator"
	"github.com/yourname/multiagent-social/internal/persistence"
	"github.com/yourname/multiagent-social/internal/pubsub"
	api "github.com/yourname/multiagent-social/internal/api"
	"github.com/yourname/multiagent-social/internal/ws"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"strings"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	addr := ":8080"
	if a := os.Getenv("HTTP_ADDR"); a != "" {
		addr = a
	}

	// Postgres DSN expected in PG_DSN env, otherwise use default local
	pgDsn := os.Getenv("PG_DSN")
	if pgDsn == "" {
		pgDsn = "postgres://postgres:password@localhost:5432/multiagent?sslmode=disable"
	}

	store, err := persistence.NewPostgresStore(ctx, pgDsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	// start pubsub (redis)
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	ps, err := pubsub.NewRedisPubSub(redisAddr)
	if err != nil {
		log.Fatalf("failed to start redis pubsub: %v", err)
	}

	orch := orchestrator.NewOrchestrator(store, ps)

	mux := http.NewServeMux()
	// health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	// metrics
	mux.Handle("/metrics", promhttp.Handler())

	apiHandler := orchestrationAPI{store: store, orchestrator: orch}
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiHandler.Router()))

	// websocket simple path (we use path prefix and let ws handler parse id)
	mux.HandleFunc("/ws/conversations/", ws.HandleConversationWS(orch, ps, store))

	// serve admin static files
	fs := http.FileServer(http.Dir("./web/admin"))
	mux.Handle("/admin/", http.StripPrefix("/admin/", fs))

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Printf("server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
	fmt.Println("server stopped")
}

// orchestrationAPI wires a minimal set of handlers for MVP.
type orchestrationAPI struct {
	store        *persistence.PostgresStore
	orchestrator *orchestrator.Orchestrator
}

func (a orchestrationAPI) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			a.listAgents(w, r)
			return
		}
		if r.Method == http.MethodPost {
			// require admin
			api.RequireAdmin(http.HandlerFunc(a.createAgent)).ServeHTTP(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			a.listConversations(w, r)
			return
		}
		if r.Method == http.MethodPost {
			a.createConversation(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// nested conv routes: /conversations/{id}/...
	mux.HandleFunc("/conversations/", func(w http.ResponseWriter, r *http.Request) {
		trim := strings.TrimPrefix(r.URL.Path, "/conversations/")
		trim = strings.Trim(trim, "/")
		parts := strings.Split(trim, "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		id := parts[0]
		if len(parts) >= 2 && parts[1] == "messages" {
			if r.Method == http.MethodPost {
				// attach id to URL for handler compatibility
				r = r.WithContext(context.WithValue(r.Context(), "convID", id))
				a.postMessage(w, r)
				return
			}
		}
		if len(parts) >= 2 && parts[1] == "debate" {
			if r.Method == http.MethodPost {
				r = r.WithContext(context.WithValue(r.Context(), "convID", id))
				a.startDebate(w, r)
				return
			}
		}
		http.Error(w, "not found", http.StatusNotFound)
	})

	return mux
}

func (a orchestrationAPI) listAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := a.store.ListAgents(r.Context())
	if err != nil {
		http.Error(w, "failed to list agents", http.StatusInternalServerError)
		return
	}
	_ = agents
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(agents)
}

func (a orchestrationAPI) createAgent(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name            string                 `json:"name"`
		Persona         string                 `json:"persona"`
		BehaviorProfile map[string]interface{} `json:"behavior_profile"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	id, err := a.store.CreateAgent(r.Context(), payload.Name, payload.Persona, payload.BehaviorProfile)
	if err != nil {
		http.Error(w, "failed to create agent", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (a orchestrationAPI) createConversation(w http.ResponseWriter, r *http.Request) {
	// stub: create conversation with no participants
	id, err := a.orchestrator.CreateConversation(r.Context(), "Conversation (MVP)", nil)
	if err != nil {
		http.Error(w, "failed to create conversation", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(id))
}

func (a orchestrationAPI) listConversations(w http.ResponseWriter, r *http.Request) {
	list, err := a.store.ListConversations(r.Context())
	if err != nil {
		http.Error(w, "failed to list conversations", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (a orchestrationAPI) postMessage(w http.ResponseWriter, r *http.Request) {
	// convID may be provided in context by the Router helper
	convID, _ := r.Context().Value("convID").(string)
	if convID == "" {
		http.Error(w, "missing conversation id", http.StatusBadRequest)
		return
	}
	// In MVP we accept raw body as message content
	defer r.Body.Close()
	buf := make([]byte, 4096)
	n, _ := r.Body.Read(buf)
	content := string(buf[:n])
	if err := a.orchestrator.HandleUserMessage(r.Context(), convID, "user-mvp", content); err != nil {
		http.Error(w, "failed to handle message", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (a orchestrationAPI) startDebate(w http.ResponseWriter, r *http.Request) {
	convID, _ := r.Context().Value("convID").(string)
	if convID == "" {
		http.Error(w, "missing conversation id", http.StatusBadRequest)
		return
	}
	var payload struct {
		Participants []string `json:"participants"`
		Rounds       int      `json:"rounds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := a.orchestrator.StartDebate(r.Context(), convID, payload.Participants, payload.Rounds); err != nil {
		http.Error(w, "failed to start debate:"+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

