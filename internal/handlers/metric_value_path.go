package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricGetPathService описывает сервис для получения метрики по ID,
// где ID и тип метрики передаются через параметры URL.
type MetricGetPathService interface {
	// Get возвращает метрику по её ID или ошибку.
	Get(ctx context.Context, metricID types.MetricID) (*types.Metrics, error)
}

// NewMetricGetPathHandler создаёт HTTP-обработчик для получения метрики,
// где ID и тип метрики берутся из параметров URL.
// Возвращает значение метрики в виде строки или ошибку.
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

// newMetricStringValue конвертирует метрику в строковое представление
// для типа счетчика (Counter) или показателя (Gauge).
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
