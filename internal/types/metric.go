package types

import (
	"errors"
)

type MetricUpdatePathRequest struct {
	Type  MetricType
	Name  string
	Value string
}

type MetricID struct {
	ID   string     `json:"id"`
	Type MetricType `json:"type"`
}

type Metrics struct {
	MetricID
	Value *float64 `json:"value,omitempty"`
	Delta *int64   `json:"delta,omitempty"`
}

type MetricType string

const (
	CounterMetricType MetricType = "counter"
	GaugeMetricType   MetricType = "gauge"
)

var (
	ErrMetricIsNotUpdated = errors.New("metric is not updated")
)
