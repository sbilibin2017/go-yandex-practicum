package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sbilibin2017/go-yandex-practicum/internal/errors"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricUpdatesBodyService interface {
	Updates(ctx context.Context, metrics []*types.Metrics) ([]*types.Metrics, error)
}

type MetricUpdatesBodyHandler struct {
	svc MetricUpdatesBodyService
	val func(metrics *types.Metrics) error
}

func NewMetricUpdatesBodyHandler(
	svc MetricUpdatesBodyService,
	val func(metrics *types.Metrics) error,
) *MetricUpdatesBodyHandler {
	return &MetricUpdatesBodyHandler{
		svc: svc,
		val: val,
	}
}

func (h *MetricUpdatesBodyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var metrics []*types.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, m := range metrics {
		if err := h.val(m); err != nil {
			handleMetricUpdatesBodyError(w, err)
			return
		}
	}
	_, err := h.svc.Updates(r.Context(), metrics)
	if err != nil {
		handleMetricUpdatesBodyError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleMetricUpdatesBodyError(w http.ResponseWriter, err error) {
	switch err {
	case errors.ErrMetricIDRequired:
		w.WriteHeader(http.StatusNotFound)
	case errors.ErrUnsupportedMetricType,
		errors.ErrCounterValueRequired,
		errors.ErrGaugeValueRequired:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
