package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Router returns a subrouter for API handlers (placeholder for MVP).
func Router() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return r
}

