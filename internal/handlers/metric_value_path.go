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

func NewMetricGetPathHandler(
	svc MetricGetPathService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `urlparam:"name"`
			Type string `urlparam:"type"`
		}
		parseURLParam(r, &req)

		var metricID types.MetricID
		metricID.ID = req.Name
		if (req.Type != string(types.CounterMetricType)) && (req.Type != string(types.GaugeMetricType)) {
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
		metricID.Type = types.MetricType(req.Type)

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

		value := newMetricStringValue(*metric)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(value))
	}
}

func newMetricStringValue(m types.Metrics) string {
	var value string
	if m.Type == types.CounterMetricType {
		if m.Delta != nil {
			value = strconv.FormatInt(*m.Delta, 10)
		}
	} else if m.Type == types.GaugeMetricType {
		if m.Value != nil {
			value = strconv.FormatFloat(*m.Value, 'f', -1, 64)
		}
	}
	return value
}
