package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// PingHandlerConfig holds configuration for the PingDBHandler, including the database connection.
type PingHandlerConfig struct {
	DB *sqlx.DB
}

// PingHandlerOption defines a functional option for configuring PingHandlerConfig.
type PingHandlerOption func(*PingHandlerConfig)

// NewPingHandlerConfig creates a new PingHandlerConfig applying the provided options.
func NewPingHandlerConfig(opts ...PingHandlerOption) *PingHandlerConfig {
	cfg := &PingHandlerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// WithPingDB sets the database connection in the PingHandlerConfig.
func WithPingDB(db *sqlx.DB) PingHandlerOption {
	return func(cfg *PingHandlerConfig) {
		cfg.DB = db
	}
}

// PingDBHandler handles HTTP requests for database connectivity check (ping).
type PingDBHandler struct {
	config *PingHandlerConfig
}

// NewPingDBHandler creates a new PingDBHandler with the given options.
func NewPingDBHandler(opts ...PingHandlerOption) *PingDBHandler {
	return &PingDBHandler{config: NewPingHandlerConfig(opts...)}
}

// serveHTTP responds to /ping requests by checking database connectivity.
func (h *PingDBHandler) serveHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	if h.config.DB == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	if err := h.config.DB.PingContext(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RegisterRoute registers the /ping route on the provided router.
func (h *PingDBHandler) RegisterRoute(r chi.Router) {
	r.Get("/ping", h.serveHTTP)
}
