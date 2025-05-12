package handlers

import (
	"context"
	"fmt"
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

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		renderMetricsHTML(w, metrics)
	}
}

func renderMetricsHTML(w http.ResponseWriter, metrics []types.Metrics) {
	fmt.Fprintln(w, "<!DOCTYPE html><html><head><title>Metrics</title></head><body><ul>")
	for _, m := range metrics {
		switch m.Type {
		case types.GaugeMetricType:
			if m.Value != nil {
				fmt.Fprintf(w, "<li>%s: %v</li>\n", m.ID, *m.Value)
			}
		case types.CounterMetricType:
			if m.Delta != nil {
				fmt.Fprintf(w, "<li>%s: %d</li>\n", m.ID, *m.Delta)
			}
		}
	}
	fmt.Fprintln(w, "</ul></body></html>")
}
