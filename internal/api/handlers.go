package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Router returns a minimal http.Handler for API endpoints without depending on chi.
func Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	// list agents
	mux.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// placeholder: return empty list
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]interface{}{})
	})
	// conversations prefix handler (list/create and nested routes)
	mux.HandleFunc("/conversations/", func(w http.ResponseWriter, r *http.Request) {
		// path: /conversations or /conversations/{id}/...
		trim := strings.TrimPrefix(r.URL.Path, "/conversations/")
		if trim == "" || trim == "/" {
			if r.Method == http.MethodGet {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode([]interface{}{})
				return
			}
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte("conv-local"))
				return
			}
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		parts := strings.Split(strings.Trim(trim, "/"), "/")
		id := parts[0]
		_ = id
		if len(parts) >= 2 && parts[1] == "messages" {
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte("accepted"))
				return
			}
		}
		http.Error(w, "not found", http.StatusNotFound)
	})
	return mux
}

