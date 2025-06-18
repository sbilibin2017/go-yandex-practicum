package types

const (
	Counter = "counter"
	Gauge   = "gauge"
)

type MetricID struct {
	ID   string `json:"id" db:"id"`
	Type string `json:"type" db:"type"`
}

type Metrics struct {
	ID    string   `json:"id" db:"id"`
	Type  string   `json:"type" db:"type"`
	Value *float64 `json:"value,omitempty" db:"value"`
	Delta *int64   `json:"delta,omitempty" db:"delta"`
}
