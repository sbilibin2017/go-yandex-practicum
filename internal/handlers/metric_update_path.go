package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricUpdatePathService interface {
	Updates(ctx context.Context, metrics []*types.Metrics) ([]*types.Metrics, error)
}

func NewMetricUpdatePathHandler(svc MetricUpdatePathService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		mType := r.URL.Query().Get("type")
		value := r.URL.Query().Get("value")

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
		case string(types.Counter):
			delta, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			metric.Delta = &delta
			metric.Type = types.Counter
		case string(types.Gauge):
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

		_, err := svc.Updates(r.Context(), []*types.Metrics{&metric})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
