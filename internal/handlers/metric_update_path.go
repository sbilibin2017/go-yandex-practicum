package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/errors"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricUpdatePathService interface {
	Updates(ctx context.Context, metrics []*types.Metrics) ([]*types.Metrics, error)
}

type MetricUpdatePathHandler struct {
	svc MetricUpdatePathService
	val func(mType string, mName string, mValue string) error
}

func NewMetricUpdatePathHandler(
	svc MetricUpdatePathService,
	val func(mType string, mName string, mValue string) error,
) *MetricUpdatePathHandler {
	return &MetricUpdatePathHandler{
		svc: svc, val: val,
	}
}

func (h *MetricUpdatePathHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")
	mValue := chi.URLParam(r, "value")

	err := h.val(mType, mName, mValue)
	if err != nil {
		handleMetricupdatePathError(w, err)
		return
	}

	metrics := types.NewMetricFromAttributes(mType, mName, mValue)

	_, err = h.svc.Updates(r.Context(), []*types.Metrics{metrics})
	if err != nil {
		handleMetricupdatePathError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleMetricupdatePathError(w http.ResponseWriter, err error) {
	switch err {
	case errors.ErrMetricNameRequired:
		w.WriteHeader(http.StatusNotFound)
	case errors.ErrUnsupportedMetricType, errors.ErrInvalidCounterValue, errors.ErrInvalidGaugeValue:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
