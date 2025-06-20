package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sbilibin2017/go-yandex-practicum/internal/errors"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricUpdateBodyService interface {
	Updates(ctx context.Context, metrics []*types.Metrics) ([]*types.Metrics, error)
}

type MetricUpdateBodyHandler struct {
	svc MetricUpdateBodyService
	val func(metrics *types.Metrics) error
}

func NewMetricUpdateBodyHandler(
	svc MetricUpdateBodyService,
	val func(metrics *types.Metrics) error,
) *MetricUpdateBodyHandler {
	return &MetricUpdateBodyHandler{
		svc: svc,
		val: val,
	}
}

func (h *MetricUpdateBodyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var metric *types.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.val(metric); err != nil {
		handleMetricUpdateBodyError(w, err)
		return
	}

	_, err := h.svc.Updates(r.Context(), []*types.Metrics{metric})
	if err != nil {
		handleMetricUpdateBodyError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleMetricUpdateBodyError(w http.ResponseWriter, err error) {
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
