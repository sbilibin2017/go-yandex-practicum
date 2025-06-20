package types

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

type MetricID struct {
	ID    string `json:"id" db:"id"`
	MType string `json:"type" db:"type"`
}

type Metrics struct {
	ID    string   `json:"id" db:"id"`
	MType string   `json:"type" db:"type"`
	Value *float64 `json:"value,omitempty" db:"value"`
	Delta *int64   `json:"delta,omitempty" db:"delta"`
}

func NewMetricFromAttributes(mType string, mName string, mValue string) *Metrics {
	switch mType {
	case Counter:
		delta, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return nil
		}
		return &Metrics{
			ID:    mName,
			MType: Counter,
			Delta: &delta,
		}

	case Gauge:
		val, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return nil
		}
		return &Metrics{
			ID:    mName,
			MType: Gauge,
			Value: &val,
		}

	default:
		return nil
	}
}

func NewMetricsHTML(metrics []*Metrics) string {
	var builder strings.Builder
	builder.WriteString("<!DOCTYPE html><html><head><title>Metrics</title></head><body><ul>\n")

	for _, m := range metrics {
		switch m.MType {
		case Gauge:
			if m.Value != nil {
				builder.WriteString(fmt.Sprintf("<li>%s: %v</li>\n", m.ID, *m.Value))
			}
		case Counter:
			if m.Delta != nil {
				builder.WriteString(fmt.Sprintf("<li>%s: %d</li>\n", m.ID, *m.Delta))
			}
		}
	}

	builder.WriteString("</ul></body></html>\n")
	return builder.String()
}

func NewMetricString(metric *Metrics) string {
	if metric == nil {
		return ""
	}

	switch metric.MType {
	case Counter:
		if metric.Delta != nil {
			return strconv.FormatInt(*metric.Delta, 10)
		}
	case Gauge:
		if metric.Value != nil {
			return strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		}
	}

	return ""
}
