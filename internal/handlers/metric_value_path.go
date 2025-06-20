package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/errors"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricGetPathService interface {
	Get(ctx context.Context, metricID types.MetricID) (*types.Metrics, error)
}

type MetricGetPathHandler struct {
	svc MetricGetPathService
	val func(mType string, mName string) error
}

func NewMetricGetPathHandler(
	svc MetricGetPathService,
	val func(mType string, mName string) error,
) *MetricGetPathHandler {
	return &MetricGetPathHandler{
		svc: svc,
		val: val,
	}
}

func (h *MetricGetPathHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")

	err := h.val(mType, mName)
	if err != nil {
		handleMetricGetPathError(w, err)
		return
	}

	metricID := types.MetricID{
		ID:    mName,
		MType: mType,
	}

	metric, err := h.svc.Get(r.Context(), metricID)
	if err != nil {
		handleMetricGetPathError(w, err)
		return
	}

	if metric == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(types.NewMetricString(metric)))
}

func handleMetricGetPathError(w http.ResponseWriter, err error) {
	switch err {
	case errors.ErrMetricNameRequired:
		w.WriteHeader(http.StatusNotFound)
	case errors.ErrUnsupportedMetricType:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
