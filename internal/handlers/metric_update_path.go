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

func NewMetricUpdatePathHandler(
	svc MetricUpdatePathService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MetricUpdatePathRequest
		parseURLParam(r, &req)

		var metric types.Metrics
		if !newMetricFromMetricUpdatePathRequest(w, req, &metric) {
			return
		}

		if err := svc.Update(r.Context(), []types.Metrics{metric}); err != nil {
			http.Error(w, "Metric not updated", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Metric updated successfully"))
	}
}

func newMetricFromMetricUpdatePathRequest(
	w http.ResponseWriter,
	req types.MetricUpdatePathRequest,
	metric *types.Metrics,
) bool {
	if req.Name == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return false
	}
	metric.ID = req.Name

	if req.Value == "" {
		http.Error(w, "Metric value is required", http.StatusBadRequest)
		return false
	}

	switch req.Type {
	case string(types.CounterMetricType):
		delta, err := strconv.ParseInt(req.Value, 10, 64)
		if err != nil {
			http.Error(w, "Invalid metric value for counter", http.StatusBadRequest)
			return false
		}
		metric.Delta = &delta
		metric.Type = types.CounterMetricType
	case string(types.GaugeMetricType):
		value, err := strconv.ParseFloat(req.Value, 64)
		if err != nil {
			http.Error(w, "Invalid metric value for gauge", http.StatusBadRequest)
			return false
		}
		metric.Value = &value
		metric.Type = types.GaugeMetricType
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return false
	}

	return true
}
