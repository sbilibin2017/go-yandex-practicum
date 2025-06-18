package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricGetPathService interface {
	Get(ctx context.Context, metricID types.MetricID) (*types.Metrics, error)
}

func NewMetricGetPathHandler(svc MetricGetPathService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		mType := chi.URLParam(r, "type")

		if id == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if mType != types.Counter && mType != types.Gauge {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metricID := types.MetricID{
			ID:   id,
			Type: mType,
		}

		metric, err := svc.Get(r.Context(), metricID)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if metric == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var value string
		if metric.Type == types.Counter {
			if metric.Delta != nil {
				value = strconv.FormatInt(*metric.Delta, 10)
			}
		} else if metric.Type == types.Gauge {
			if metric.Value != nil {
				value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(value))
	}
}
