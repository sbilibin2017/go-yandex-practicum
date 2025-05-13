package types

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type MetricType string

const (
	CounterMetricType MetricType = "counter"
	GaugeMetricType   MetricType = "gauge"
)

type MetricID struct {
	ID   string     `json:"id"`
	Type MetricType `json:"type"`
}

type Metrics struct {
	MetricID
	Value *float64 `json:"value,omitempty"`
	Delta *int64   `json:"delta,omitempty"`
}

var (
	ErrMetricInternal = errors.New("internal error")
	ErrMetricNotFound = errors.New("metric not found")
)

func NewMetricStringValue(m Metrics) string {
	var value string
	if m.Type == CounterMetricType {
		if m.Delta != nil {
			value = strconv.FormatInt(*m.Delta, 10)
		}
	} else if m.Type == GaugeMetricType {
		if m.Value != nil {
			value = strconv.FormatFloat(*m.Value, 'f', -1, 64)
		}
	}
	return value
}

func NewMetricsHTML(metrics []Metrics) string {
	var builder strings.Builder
	builder.WriteString("<!DOCTYPE html><html><head><title>Metrics</title></head><body><ul>\n")
	for _, m := range metrics {
		switch m.Type {
		case GaugeMetricType:
			if m.Value != nil {
				builder.WriteString(fmt.Sprintf("<li>%s: %v</li>\n", m.ID, *m.Value))
			}
		case CounterMetricType:
			if m.Delta != nil {
				builder.WriteString(fmt.Sprintf("<li>%s: %d</li>\n", m.ID, *m.Delta))
			}
		}
	}
	builder.WriteString("</ul></body></html>\n")
	return builder.String()
}
