package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricUpdatesBodyService interface {
	Updates(ctx context.Context, metrics []types.Metrics) ([]types.Metrics, error)
}

func NewMetricUpdatesBodyHandler(
	svc MetricUpdatesBodyService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reqs []types.Metrics

		if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		for _, req := range reqs {
			if req.ID == "" {
				http.Error(w, "Metric id is required", http.StatusNotFound)
				return
			}

			switch req.Type {
			case types.CounterMetricType:
				if req.Delta == nil {
					http.Error(w, "Metric delta is required for counter", http.StatusBadRequest)
					return
				}
			case types.GaugeMetricType:
				if req.Value == nil {
					http.Error(w, "Metric value is required for gauge", http.StatusBadRequest)
					return
				}
			default:
				http.Error(w, "Invalid metric type", http.StatusBadRequest)
				return
			}

		}

		metricsUpdated, err := svc.Updates(r.Context(), reqs)
		if err != nil {
			http.Error(w, "Metric not updated", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(metricsUpdated); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
