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

	"github.com/go-chi/chi/v5"

	"github.com/yourname/multiagent-social/internal/orchestrator"
	"github.com/yourname/multiagent-social/internal/persistence"
	"github.com/yourname/multiagent-social/internal/pubsub"
	api "github.com/yourname/multiagent-social/internal/api"
	"github.com/yourname/multiagent-social/internal/ws"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	// metrics
	r.Handle("/metrics", promhttp.Handler())

	api := orchestrationAPI{store: store, orchestrator: orch}
	r.Mount("/api/v1", api.Router())
	// websocket endpoint for conversations
	r.Route("/ws", func(sr chi.Router) {
		sr.Get("/conversations/{id}", ws.HandleConversationWS(orch, ps, store))
	})
	// serve admin static files
	fs := http.FileServer(http.Dir("./web/admin"))
	r.Handle("/admin/*", http.StripPrefix("/admin/", fs))

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
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
	r := chi.NewRouter()
	r.Get("/agents", a.listAgents)
	// protect agent creation with admin JWT
	r.With(api.RequireAdmin).Post("/agents", a.createAgent)
	r.Post("/conversations", a.createConversation)
	r.Post("/conversations/{id}/messages", a.postMessage)
	r.Post("/conversations/{id}/debate", a.startDebate)
	r.Get("/conversations", a.listConversations)
	return r
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
	convID := chi.URLParam(r, "id")
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
	convID := chi.URLParam(r, "id")
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

