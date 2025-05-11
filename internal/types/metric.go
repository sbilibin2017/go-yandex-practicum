package types

import (
	"errors"
)

type Metrics struct {
	ID    string
	Type  string
	Value float64
	Delta int64
}

type MetricID struct {
	ID   string
	Type string
}

const (
	CounterMetricType string = "counter"
	GaugeMetricType   string = "gauge"
)

var (
	ErrInternal       = errors.New("internal error")
	ErrMetricNotFound = errors.New("metric not found")
)
