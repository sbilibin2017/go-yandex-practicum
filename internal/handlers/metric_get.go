package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricGetter defines the interface for retrieving a metric by ID.
type MetricGetter interface {
	Get(ctx context.Context, metricID types.MetricID) (*types.Metrics, error)
}

// MetricGetterHandlerConfig holds the configuration for metric getter handlers.
type MetricGetterHandlerConfig struct {
	svc MetricGetter
}

// MetricGetterHandlerOption is a functional option for configuring a MetricGetterHandlerConfig.
type MetricGetterHandlerOption func(*MetricGetterHandlerConfig)

// WithMetricGetter sets the MetricGetter service for the handler configuration.
func WithMetricGetter(svc MetricGetter) MetricGetterHandlerOption {
	return func(cfg *MetricGetterHandlerConfig) {
		cfg.svc = svc
	}
}

// NewMetricGetterHandlerConfig creates a new MetricGetterHandlerConfig using the provided options.
func NewMetricGetterHandlerConfig(opts ...MetricGetterHandlerOption) *MetricGetterHandlerConfig {
	cfg := &MetricGetterHandlerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// MetricGetPathHandler handles metric retrieval via HTTP GET request using URL parameters.
type MetricGetPathHandler struct {
	svc MetricGetter
}

// NewMetricGetPathHandler constructs a MetricGetPathHandler using provided options.
func NewMetricGetPathHandler(opts ...MetricGetterHandlerOption) *MetricGetPathHandler {
	cfg := NewMetricGetterHandlerConfig(opts...)
	return &MetricGetPathHandler{svc: cfg.svc}
}

// serveHTTP processes GET requests to retrieve a metric by type and name in the URL path.
func (h *MetricGetPathHandler) serveHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	if name == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if mType != types.Counter && mType != types.Gauge {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricID := types.MetricID{ID: name, MType: mType}
	metric, err := h.svc.Get(r.Context(), metricID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if metric == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var valueString string
	switch mType {
	case types.Counter:
		if metric.Delta == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		valueString = strconv.FormatInt(*metric.Delta, 10)
	case types.Gauge:
		if metric.Value == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		valueString = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(valueString))
}

// RegisterRoute registers HTTP GET endpoints for retrieving metrics by path.
func (h *MetricGetPathHandler) RegisterRoute(r chi.Router) {
	r.Get("/value/{type}/{name}", h.serveHTTP)
	r.Get("/value/{type}", h.serveHTTP)
}

// MetricGetBodyHandler handles metric retrieval via HTTP POST request using JSON in the body.
type MetricGetBodyHandler struct {
	svc MetricGetter
}

// NewMetricGetBodyHandler constructs a MetricGetBodyHandler using provided options.
func NewMetricGetBodyHandler(opts ...MetricGetterHandlerOption) *MetricGetBodyHandler {
	cfg := NewMetricGetterHandlerConfig(opts...)
	return &MetricGetBodyHandler{svc: cfg.svc}
}

// serveHTTP processes POST requests to retrieve a metric specified in the JSON request body.
func (h *MetricGetBodyHandler) serveHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var metricID types.MetricID
	if err := json.NewDecoder(r.Body).Decode(&metricID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if metricID.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if metricID.MType != types.Counter && metricID.MType != types.Gauge {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := h.svc.Get(r.Context(), metricID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if metric == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metric)
}

// RegisterRoute registers HTTP POST endpoint for retrieving a metric by body.
func (h *MetricGetBodyHandler) RegisterRoute(r chi.Router) {
	r.Post("/value/", h.serveHTTP)
}
