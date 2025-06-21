package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// PingDBHandler handles HTTP requests for database connectivity check (ping).
type PingDBHandler struct {
	db *sqlx.DB
}

// PingHandlerOption defines a functional option for configuring PingDBHandler.
type PingHandlerOption func(*PingDBHandler)

// WithPingDB sets the database connection on the PingDBHandler.
func WithPingDB(db *sqlx.DB) PingHandlerOption {
	return func(h *PingDBHandler) {
		h.db = db
	}
}

// NewPingDBHandler creates a new PingDBHandler with the given options.
func NewPingDBHandler(opts ...PingHandlerOption) *PingDBHandler {
	h := &PingDBHandler{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// serveHTTP responds to /ping requests by checking database connectivity.
func (h *PingDBHandler) serveHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	if h.db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	if err := h.db.PingContext(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RegisterRoute registers the /ping route on the provided router.
func (h *PingDBHandler) RegisterRoute(r chi.Router) {
	r.Get("/ping", h.serveHTTP)
}
