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

func NewMetricGetBodyHandler(svc MetricGetBodyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MetricID

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if req.ID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.Type != types.Counter && req.Type != types.Gauge {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metric, err := svc.Get(r.Context(), req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if metric == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(metric); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
