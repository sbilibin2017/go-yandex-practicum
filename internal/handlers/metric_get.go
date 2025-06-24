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

// --- Functional Options for MetricGetPathHandler ---

type MetricGetPathHandler struct {
	svc MetricGetter
}

type MetricGetPathHandlerOption func(*MetricGetPathHandler)

// WithMetricGetter sets the MetricGetter service on MetricGetPathHandler.
func WithMetricGetterPath(svc MetricGetter) MetricGetPathHandlerOption {
	return func(h *MetricGetPathHandler) {
		h.svc = svc
	}
}

func NewMetricGetPathHandler(opts ...MetricGetPathHandlerOption) *MetricGetPathHandler {
	h := &MetricGetPathHandler{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *MetricGetPathHandler) serveHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	Type := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	if name == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if Type != types.Counter && Type != types.Gauge {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricID := types.MetricID{ID: name, Type: Type}
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
	switch Type {
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

func (h *MetricGetPathHandler) RegisterRoute(r chi.Router) {
	r.Get("/value/{type}/{name}", h.serveHTTP)
	r.Get("/value/{type}", h.serveHTTP)
}

// --- Functional Options for MetricGetBodyHandler ---

type MetricGetBodyHandler struct {
	svc MetricGetter
}

type MetricGetBodyHandlerOption func(*MetricGetBodyHandler)

// WithMetricGetter sets the MetricGetter service on MetricGetBodyHandler.
func WithMetricGetterBody(svc MetricGetter) MetricGetBodyHandlerOption {
	return func(h *MetricGetBodyHandler) {
		h.svc = svc
	}
}

func NewMetricGetBodyHandler(opts ...MetricGetBodyHandlerOption) *MetricGetBodyHandler {
	h := &MetricGetBodyHandler{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

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

	if metricID.Type != types.Counter && metricID.Type != types.Gauge {
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

func (h *MetricGetBodyHandler) RegisterRoute(r chi.Router) {
	r.Post("/value/", h.serveHTTP)
}
