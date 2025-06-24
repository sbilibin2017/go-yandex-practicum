package types

const (
	Counter = "counter" // Counter represents a metric that only increments (integer).
	Gauge   = "gauge"   // Gauge represents a metric that can hold arbitrary float64 values.
)

type MetricID struct {
	ID   string `json:"id" db:"id"`     // ID is the unique identifier/name of the metric.
	Type string `json:"type" db:"type"` // Type specifies the metric type (e.g., "counter", "gauge").
}

type Metrics struct {
	ID    string   `json:"id" db:"id"`                 // ID is the unique identifier/name of the metric.
	Type  string   `json:"type" db:"type"`             // Type specifies the metric type (e.g., "counter", "gauge").
	Value *float64 `json:"value,omitempty" db:"value"` // Value is used for gauge metrics (float64), nil for counters.
	Delta *int64   `json:"delta,omitempty" db:"delta"` // Delta is used for counter metrics (int64), nil for gauges.
}
