package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricUpdatePathService interface {
	Update(ctx context.Context, metrics []types.Metrics) error
}

func MetricUpdatePathHandler(svc MetricUpdatePathService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mtype := getURLParam(r, "type")
		name := getURLParam(r, "name")
		value := getURLParam(r, "value")

		if mtype != types.CounterMetricType && mtype != types.GaugeMetricType {
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		if name == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		if value == "" {
			http.Error(w, "Metric value is required", http.StatusBadRequest)
			return
		}

		var metrics []types.Metrics

		if mtype == types.CounterMetricType {
			delta, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				http.Error(w, "Invalid metric value for counter", http.StatusBadRequest)
				return
			}
			metrics = append(metrics, types.Metrics{
				ID:    name,
				Type:  mtype,
				Delta: delta,
			})
		} else if mtype == types.GaugeMetricType {
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				http.Error(w, "Invalid metric value for gauge", http.StatusBadRequest)
				return
			}
			metrics = append(metrics, types.Metrics{
				ID:    name,
				Type:  mtype,
				Value: val,
			})
		}

		if err := svc.Update(r.Context(), metrics); err != nil {
			http.Error(w, "Metric not updated", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Metric updated successfully"))
	}
}
