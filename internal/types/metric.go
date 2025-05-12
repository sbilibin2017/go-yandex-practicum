package types

import (
	"errors"
)

type MetricUpdatePathRequest struct {
	Name  string `urlparam:"name"`
	Type  string `urlparam:"type"`
	Value string `urlparam:"value"`
}

type MetricGetPathRequest struct {
	Name string `urlparam:"name"`
	Type string `urlparam:"type"`
}

type Metrics struct {
	MetricID
	Value *float64
	Delta *int64
}

func NewMetrics(id string, mtype string, delta *int64, value *float64) *Metrics {
	return &Metrics{
		MetricID: MetricID{
			ID:   id,
			Type: MetricType(mtype),
		},
		Delta: delta,
		Value: value,
	}
}

type MetricID struct {
	ID   string
	Type MetricType
}

type MetricType string

const (
	CounterMetricType MetricType = "counter"
	GaugeMetricType   MetricType = "gauge"
)

var (
	ErrMetricInternal = errors.New("internal error")
	ErrMetricNotFound = errors.New("metric not found")
)
