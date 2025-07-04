package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	pb "github.com/sbilibin2017/go-yandex-practicum/protos"
)

// MetricUpdater defines an interface for updating multiple metrics.
type MetricUpdater interface {
	Updates(ctx context.Context, metrics []*types.Metrics) ([]*types.Metrics, error)
}

// Functional options for MetricUpdatePathHandler
type MetricUpdatePathHandlerOption func(*MetricUpdatePathHandler)

func WithMetricUpdaterPath(svc MetricUpdater) MetricUpdatePathHandlerOption {
	return func(h *MetricUpdatePathHandler) {
		h.svc = svc
	}
}

// MetricUpdatePathHandler handles metric updates via URL path parameters.
type MetricUpdatePathHandler struct {
	svc MetricUpdater
}

func NewMetricUpdatePathHandler(opts ...MetricUpdatePathHandlerOption) *MetricUpdatePathHandler {
	h := &MetricUpdatePathHandler{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *MetricUpdatePathHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	Type := chi.URLParam(r, "type")
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

	switch Type {
	case types.Counter:
		delta, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		metric.Delta = &delta
		metric.Type = types.Counter

	case types.Gauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		metric.Value = &val
		metric.Type = types.Gauge

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

func (h *MetricUpdatePathHandler) RegisterRoute(r chi.Router) {
	r.Post("/update/{type}/{name}/{value}", h.Update)
	r.Post("/update/{type}/{name}", h.Update)
}

// Functional options for MetricUpdateBodyHandler
type MetricUpdateBodyHandlerOption func(*MetricUpdateBodyHandler)

func WithMetricUpdaterBody(svc MetricUpdater) MetricUpdateBodyHandlerOption {
	return func(h *MetricUpdateBodyHandler) {
		h.svc = svc
	}
}

// MetricUpdateBodyHandler handles metric updates sent via JSON request body.
type MetricUpdateBodyHandler struct {
	svc MetricUpdater
}

func NewMetricUpdateBodyHandler(opts ...MetricUpdateBodyHandlerOption) *MetricUpdateBodyHandler {
	h := &MetricUpdateBodyHandler{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *MetricUpdateBodyHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	switch metric.Type {
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

func (h *MetricUpdateBodyHandler) RegisterRoute(r chi.Router) {
	r.Post("/update/", h.Update)
}

// Functional options for MetricUpdatesBodyHandler
type MetricUpdatesBodyHandlerOption func(*MetricUpdatesBodyHandler)

func WithMetricUpdaterBatchBody(svc MetricUpdater) MetricUpdatesBodyHandlerOption {
	return func(h *MetricUpdatesBodyHandler) {
		h.svc = svc
	}
}

// MetricUpdatesBodyHandler handles batch metric updates sent via JSON array.
type MetricUpdatesBodyHandler struct {
	svc MetricUpdater
}

func NewMetricUpdatesBodyHandler(opts ...MetricUpdatesBodyHandlerOption) *MetricUpdatesBodyHandler {
	h := &MetricUpdatesBodyHandler{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *MetricUpdatesBodyHandler) Updates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var metrics []*types.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(metrics) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, m := range metrics {
		if m.ID == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch m.Type {
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

func (h *MetricUpdatesBodyHandler) RegisterRoute(r chi.Router) {
	r.Post("/updates/", h.Updates)
}

// MetricUpdaterService implements the gRPC server interface.
type MetricGRPCUpdaterHandler struct {
	pb.UnimplementedMetricUpdaterServer
	updater MetricUpdater
}

func NewMetricGRPCUpdaterHandler(updater MetricUpdater) *MetricGRPCUpdaterHandler {
	return &MetricGRPCUpdaterHandler{
		updater: updater,
	}
}

// Convert pb.Metric to types.Metrics
func toMetric(m *pb.Metric) *types.Metrics {
	v := m.GetValue()
	d := m.GetDelta()
	return &types.Metrics{
		ID:    m.GetId(),
		Type:  m.GetType(),
		Value: &v,
		Delta: &d,
	}
}

// Convert types.Metrics to pb.Metric
func fromMetric(m *types.Metrics) *pb.Metric {
	return &pb.Metric{
		Id:    m.ID,
		Type:  m.Type,
		Value: *m.Value,
		Delta: *m.Delta,
	}
}

// Updates implements the gRPC server method, adapting calls to your internal interface.
func (s *MetricGRPCUpdaterHandler) Updates(
	ctx context.Context,
	req *pb.UpdateMetricsRequest,
) (*pb.UpdateMetricsResponse, error) {
	metrics := make([]*types.Metrics, 0, len(req.GetMetrics()))
	for _, m := range req.GetMetrics() {
		metrics = append(metrics, toMetric(m))
	}

	updatedMetrics, err := s.updater.Updates(ctx, metrics)
	if err != nil {
		return &pb.UpdateMetricsResponse{
			Error: err.Error(),
		}, nil
	}

	respMetrics := make([]*pb.Metric, 0, len(updatedMetrics))
	for _, m := range updatedMetrics {
		respMetrics = append(respMetrics, fromMetric(m))
	}

	return &pb.UpdateMetricsResponse{
		Metrics: respMetrics,
	}, nil
}
