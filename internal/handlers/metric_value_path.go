package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricGetPathService interface {
	Get(ctx context.Context, metricID types.MetricID) (*types.Metrics, error)
}

func MetricGetPathHandler(
	svc MetricGetPathService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MetricGetPathRequest
		parseURLParam(r, &req)

		var metricID types.MetricID
		if !newMetricIDFromMetricGetPathRequest(w, req, &metricID) {
			return
		}

		metric, err := svc.Get(r.Context(), metricID)

		if err != nil {
			switch err {
			case types.ErrMetricNotFound:
				http.Error(w, "Metric not found", http.StatusNotFound)
			default:
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		value := convertMetricToString(metric)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(value))
	}
}

func newMetricIDFromMetricGetPathRequest(
	w http.ResponseWriter,
	req types.MetricGetPathRequest,
	metric *types.MetricID,
) bool {
	metric.ID = req.Name
	if (req.Type != string(types.CounterMetricType)) && (req.Type != string(types.GaugeMetricType)) {
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return false
	}
	metric.Type = types.MetricType(req.Type)
	return true
}

func convertMetricToString(metric *types.Metrics) string {
	var value string

	if metric.Type == types.CounterMetricType {
		if metric.Delta != nil {
			value = strconv.FormatInt(*metric.Delta, 10)
		}
	} else if metric.Type == types.GaugeMetricType {
		if metric.Value != nil {
			value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		}
	}
	return value
}
