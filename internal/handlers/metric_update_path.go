package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricUpdatePathService interface {
	Update(ctx context.Context, metrics []types.Metrics) error
}

func MetricUpdatePathHandler(svc MetricUpdatePathService) http.Handler {
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		metricValue := chi.URLParam(r, "value")

		if metricType == "" {
			http.Error(w, "Metric type is required", http.StatusBadRequest)
			return
		}

		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		if metricValue == "" {
			http.Error(w, "Metric value is required", http.StatusBadRequest)
			return
		}

		var metrics []types.Metrics

		if metricType == string(types.CounterMetricType) {
			delta, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid metric value for counter", http.StatusBadRequest)
				return
			}
			metrics = append(metrics, types.Metrics{
				MetricID: types.MetricID{
					ID:   metricName,
					Type: types.CounterMetricType,
				},
				Delta: &delta,
			})
		} else if metricType == string(types.GaugeMetricType) {
			val, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Invalid metric value for gauge", http.StatusBadRequest)
				return
			}
			metrics = append(metrics, types.Metrics{
				MetricID: types.MetricID{
					ID:   metricName,
					Type: types.GaugeMetricType,
				},
				Value: &val,
			})
		} else {
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		if err := svc.Update(r.Context(), metrics); err != nil {
			http.Error(w, "Metric not updated", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Metric updated successfully"))
	})
	return r
}
