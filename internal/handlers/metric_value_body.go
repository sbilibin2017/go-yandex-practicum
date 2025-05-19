package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricGetBodyService описывает сервис для получения метрики по ID,
// где ID передаётся в теле запроса.
type MetricGetBodyService interface {
	// Get возвращает метрику по её ID или ошибку.
	Get(ctx context.Context, metricID types.MetricID) (*types.Metrics, error)
}

// NewMetricGetBodyHandler создаёт HTTP-обработчик для получения метрики,
// где ID и тип метрики принимаются из JSON-тела POST-запроса.
// Возвращает метрику в JSON или ошибку.
func NewMetricGetBodyHandler(
	svc MetricGetBodyService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MetricID

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if req.ID == "" {
			http.Error(w, "Metric ID is required", http.StatusBadRequest)
			return
		}
		if req.Type != types.CounterMetricType && req.Type != types.GaugeMetricType {
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		metric, err := svc.Get(r.Context(), req)
		if err != nil {
			switch err {
			case types.ErrMetricNotFound:
				http.Error(w, "Metric not found", http.StatusNotFound)
			default:
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(metric); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
