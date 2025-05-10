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
		var req types.MetricUpdatePathRequest
		req.Type = types.MetricType(chi.URLParam(r, "type"))
		req.Name = chi.URLParam(r, "name")
		req.Value = chi.URLParam(r, "value")

		if req.Type == "" {
			http.Error(w, "Metric type is required", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		if req.Value == "" {
			http.Error(w, "Metric value is required", http.StatusBadRequest)
			return
		}

		var metrics []types.Metrics

		if req.Type == types.CounterMetricType {
			delta, err := strconv.ParseInt(req.Value, 10, 64)
			if err != nil {
				http.Error(w, "Invalid metric value for counter", http.StatusBadRequest)
				return
			}
			metrics = append(metrics, types.Metrics{
				MetricID: types.MetricID{
					ID:   req.Name,
					Type: types.CounterMetricType,
				},
				Delta: &delta,
			})
		} else if req.Type == types.GaugeMetricType {
			val, err := strconv.ParseFloat(req.Value, 64)
			if err != nil {
				http.Error(w, "Invalid metric value for gauge", http.StatusBadRequest)
				return
			}
			metrics = append(metrics, types.Metrics{
				MetricID: types.MetricID{
					ID:   req.Name,
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
