package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetricStringValue(t *testing.T) {
	t.Run("counter", func(t *testing.T) {
		val := int64(42)
		m := Metrics{
			MetricID: MetricID{ID: "hits", Type: CounterMetricType},
			Delta:    &val,
		}
		assert.Equal(t, "42", NewMetricStringValue(m))
	})

	t.Run("gauge", func(t *testing.T) {
		val := 99.9
		m := Metrics{
			MetricID: MetricID{ID: "load", Type: GaugeMetricType},
			Value:    &val,
		}
		assert.Equal(t, "99.9", NewMetricStringValue(m))
	})
}

func TestNewMetricsHTML(t *testing.T) {
	delta := int64(100)
	value := 45.5
	metrics := []Metrics{
		{MetricID: MetricID{ID: "requests", Type: CounterMetricType}, Delta: &delta},
		{MetricID: MetricID{ID: "cpu", Type: GaugeMetricType}, Value: &value},
	}
	html := NewMetricsHTML(metrics)
	assert.Contains(t, html, "<li>requests: 100</li>")
	assert.Contains(t, html, "<li>cpu: 45.5</li>")
	assert.Contains(t, html, "<html>")
}
