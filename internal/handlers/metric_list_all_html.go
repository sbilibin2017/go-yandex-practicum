package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricListAllHTMLService описывает сервис для получения списка всех метрик.
type MetricListAllHTMLService interface {
	ListAll(ctx context.Context) ([]types.Metrics, error)
}

// NewMetricListAllHTMLHandler создаёт HTTP-обработчик, который выводит все метрики в HTML формате.
func NewMetricListAllHTMLHandler(svc MetricListAllHTMLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := svc.ListAll(r.Context())
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Формируем HTML с помощью отдельной функции
		metricsHTML := newMetricsHTML(metrics)

		// Устанавливаем заголовок и пишем тело ответа
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metricsHTML))
	}
}

// newMetricsHTML формирует HTML-страницу с перечислением метрик.
// Метрики выводятся в виде списка <ul><li>.
// Для Gauge выводится значение Value, для Counter — значение Delta.
func newMetricsHTML(metrics []types.Metrics) string {
	var builder strings.Builder
	builder.WriteString("<!DOCTYPE html><html><head><title>Metrics</title></head><body><ul>\n")

	for _, m := range metrics {
		switch m.Type {
		case types.GaugeMetricType:
			if m.Value != nil {
				builder.WriteString(fmt.Sprintf("<li>%s: %v</li>\n", m.ID, *m.Value))
			}
		case types.CounterMetricType:
			if m.Delta != nil {
				builder.WriteString(fmt.Sprintf("<li>%s: %d</li>\n", m.ID, *m.Delta))
			}
		}
	}

	builder.WriteString("</ul></body></html>\n")
	return builder.String()
}
