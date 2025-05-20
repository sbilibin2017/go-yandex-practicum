package types

import (
	"errors"
)

// MetricType представляет тип метрики.
// Возможные значения:
//   - CounterMetricType: счётчик (increment-only).
//   - GaugeMetricType: гейдж (текущее значение).
type MetricType string

const (
	// CounterMetricType — тип метрики-счётчика.
	CounterMetricType MetricType = "counter"
	// GaugeMetricType — тип метрики-гейджа.
	GaugeMetricType MetricType = "gauge"
)

// MetricID идентифицирует метрику по уникальному ID и её типу.
type MetricID struct {
	ID   string     `json:"id" db:"id"`     // Уникальный идентификатор метрики
	Type MetricType `json:"type" db:"type"` // Тип метрики
}

// Metrics описывает метрику с её значением.
// Включает идентификатор MetricID и дополнительные поля значения.
//
// Value используется для типа Gauge, Delta — для Counter.
type Metrics struct {
	MetricID
	Value *float64 `json:"value,omitempty" db:"value"` // Значение метрики (для Gauge)
	Delta *int64   `json:"delta,omitempty" db:"delta"` // Дельта значения (для Counter)
}

var (
	// ErrMetricInternal сигнализирует об внутренней ошибке при работе с метриками.
	ErrMetricInternal = errors.New("internal error")

	// ErrMetricNotFound возвращается, если метрика не найдена.
	ErrMetricNotFound = errors.New("metric not found")
)
