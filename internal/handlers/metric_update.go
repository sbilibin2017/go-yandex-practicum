package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricUpdater defines an interface for updating multiple metrics.
type MetricUpdater interface {
	Updates(ctx context.Context, metrics []*types.Metrics) ([]*types.Metrics, error)
}

// MetricUpdaterHandlerConfig holds configuration for MetricUpdater handlers.
type MetricUpdaterHandlerConfig struct {
	Updater MetricUpdater
}

// MetricUpdaterHandlerOption defines a functional option for MetricUpdaterHandlerConfig.
type MetricUpdaterHandlerOption func(*MetricUpdaterHandlerConfig)

// WithMetricUpdater sets the MetricUpdater service in the handler configuration.
func WithMetricUpdater(updater MetricUpdater) MetricUpdaterHandlerOption {
	return func(cfg *MetricUpdaterHandlerConfig) {
		cfg.Updater = updater
	}
}

// NewMetricUpdaterHandlerConfig creates a new MetricUpdaterHandlerConfig with the given options.
func NewMetricUpdaterHandlerConfig(opts ...MetricUpdaterHandlerOption) *MetricUpdaterHandlerConfig {
	cfg := &MetricUpdaterHandlerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// MetricUpdatePathHandler handles metric updates via URL path parameters.
type MetricUpdatePathHandler struct {
	svc MetricUpdater
}

// NewMetricUpdatePathHandler constructs a MetricUpdatePathHandler with the provided options.
func NewMetricUpdatePathHandler(opts ...MetricUpdaterHandlerOption) *MetricUpdatePathHandler {
	cfg := NewMetricUpdaterHandlerConfig(opts...)
	return &MetricUpdatePathHandler{svc: cfg.Updater}
}

// serveHTTP processes HTTP requests for updating a single metric via path parameters.
func (h *MetricUpdatePathHandler) serveHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	if name == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if value == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metric types.Metrics
	metric.ID = name

	switch mType {
	case types.Counter:
		delta, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		metric.Delta = &delta
		metric.MType = types.Counter

	case types.Gauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		metric.Value = &val
		metric.MType = types.Gauge

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err := h.svc.Updates(r.Context(), []*types.Metrics{&metric}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RegisterRoute registers the routes for the MetricUpdatePathHandler.
func (h *MetricUpdatePathHandler) RegisterRoute(r chi.Router) {
	r.Post("/update/{type}/{name}/{value}", h.serveHTTP)
	r.Post("/update/{type}/{name}", h.serveHTTP)
}

// MetricUpdateBodyHandler handles metric updates sent via JSON request body.
type MetricUpdateBodyHandler struct {
	svc MetricUpdater
}

// NewMetricUpdateBodyHandler constructs a MetricUpdateBodyHandler with the provided options.
func NewMetricUpdateBodyHandler(opts ...MetricUpdaterHandlerOption) *MetricUpdateBodyHandler {
	cfg := NewMetricUpdaterHandlerConfig(opts...)
	return &MetricUpdateBodyHandler{svc: cfg.Updater}
}

// serveHTTP processes HTTP requests for updating a single metric via JSON body.
func (h *MetricUpdateBodyHandler) serveHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var metric types.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if metric.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch metric.MType {
	case types.Counter:
		if metric.Delta == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case types.Gauge:
		if metric.Value == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updatedMetrics, err := h.svc.Updates(r.Context(), []*types.Metrics{&metric})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(updatedMetrics) > 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedMetrics[0])
	}
}

// RegisterRoute registers the routes for the MetricUpdateBodyHandler.
func (h *MetricUpdateBodyHandler) RegisterRoute(r chi.Router) {
	r.Post("/update/", h.serveHTTP)
}

// MetricUpdatesBodyHandler handles batch metric updates sent via JSON array in the request body.
type MetricUpdatesBodyHandler struct {
	svc MetricUpdater
}

// NewMetricUpdatesBodyHandler constructs a MetricUpdatesBodyHandler with the provided options.
func NewMetricUpdatesBodyHandler(opts ...MetricUpdaterHandlerOption) *MetricUpdatesBodyHandler {
	cfg := NewMetricUpdaterHandlerConfig(opts...)
	return &MetricUpdatesBodyHandler{svc: cfg.Updater}
}

// serveHTTP processes HTTP requests for updating multiple metrics via JSON body.
func (h *MetricUpdatesBodyHandler) serveHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var metrics []*types.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	for _, m := range metrics {
		if m.ID == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch m.MType {
		case types.Counter:
			if m.Delta == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case types.Gauge:
			if m.Value == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	updatedMetrics, err := h.svc.Updates(r.Context(), metrics)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedMetrics)
}

// RegisterRoute registers the routes for the MetricUpdatesBodyHandler.
func (h *MetricUpdatesBodyHandler) RegisterRoute(r chi.Router) {
	r.Post("/updates/", h.serveHTTP)
}
