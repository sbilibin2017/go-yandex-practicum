package handlers

import (
	"context"
	"net/http"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricListAllHTMLService interface {
	ListAll(ctx context.Context) ([]types.Metrics, error)
}

func NewMetricListAllHTMLHandler(svc MetricListAllHTMLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := svc.ListAll(r.Context())
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		metricsHTML := types.NewMetricsHTML(metrics)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metricsHTML))
	}
}
