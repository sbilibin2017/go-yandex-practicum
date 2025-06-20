package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sbilibin2017/go-yandex-practicum/internal/errors"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricGetBodyService interface {
	Get(ctx context.Context, metricID types.MetricID) (*types.Metrics, error)
}

type MetricGetBodyHandler struct {
	svc MetricGetBodyService
	val func(id types.MetricID) error
}

func NewMetricGetBodyHandler(
	svc MetricGetBodyService,
	val func(id types.MetricID) error,
) *MetricGetBodyHandler {
	return &MetricGetBodyHandler{
		svc: svc,
		val: val,
	}
}

func (h *MetricGetBodyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var metricID types.MetricID

	if err := json.NewDecoder(r.Body).Decode(&metricID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.val(metricID); err != nil {
		handleMetricGetBodyError(w, err)
		return
	}

	metric, err := h.svc.Get(r.Context(), metricID)
	if err != nil {
		handleMetricGetBodyError(w, err)
		return
	}

	if metric == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(metric)
}

func handleMetricGetBodyError(w http.ResponseWriter, err error) {
	switch err {
	case errors.ErrMetricIDRequired,
		errors.ErrMetricNameRequired:
		w.WriteHeader(http.StatusNotFound)
	case errors.ErrUnsupportedMetricType:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
