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
