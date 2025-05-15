package types

import (
	"errors"
)

type MetricType string

const (
	CounterMetricType MetricType = "counter"
	GaugeMetricType   MetricType = "gauge"
)

type MetricID struct {
	ID   string     `json:"id" db:"id"`
	Type MetricType `json:"type" db:"type"`
}

type Metrics struct {
	MetricID
	Value *float64 `json:"value,omitempty" db:"value"`
	Delta *int64   `json:"delta,omitempty" db:"delta"`
}

var (
	ErrMetricInternal = errors.New("internal error")
	ErrMetricNotFound = errors.New("metric not found")
)
