package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricUpdatesBodyService interface {
	Updates(ctx context.Context, metrics []*types.Metrics) ([]*types.Metrics, error)
}

func NewMetricUpdatesBodyHandler(
	svc MetricUpdatesBodyService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var reqs []*types.Metrics
		if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for _, req := range reqs {
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
		}

		metricsUpdated, err := svc.Updates(r.Context(), reqs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(metricsUpdated); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
