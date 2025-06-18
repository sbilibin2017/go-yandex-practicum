package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricUpdateBodyService interface {
	Updates(ctx context.Context, metrics []*types.Metrics) ([]*types.Metrics, error)
}

func NewMetricUpdateBodyHandler(svc MetricUpdateBodyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req *types.Metrics
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.ID == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch req.Type {
		case types.Counter:
			if req.Delta == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case types.Gauge:
			if req.Value == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metricsUpdated, err := svc.Updates(r.Context(), []*types.Metrics{req})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(metricsUpdated[0]); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
