package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricUpdateBodyService interface {
	Update(ctx context.Context, metric types.Metrics) error
}

func NewMetricUpdateBodyHandler(
	svc MetricUpdateBodyService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Metrics

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

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

		if err := svc.Update(r.Context(), req); err != nil {
			http.Error(w, "Metric not updated", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(req); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
