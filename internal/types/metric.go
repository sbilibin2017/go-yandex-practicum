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

const (
	CounterMetricType string = "counter"
	GaugeMetricType   string = "gauge"
)

var (
	ErrMetricIsNotUpdated = errors.New("metric is not updated")
)
