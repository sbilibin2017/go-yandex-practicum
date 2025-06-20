package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetricString(t *testing.T) {
	intVal := int64(42)
	floatVal := 3.14

	tests := []struct {
		name   string
		input  *Metrics
		output string
	}{
		{"nil metric", nil, ""},
		{"counter with value", &Metrics{ID: "counter1", MType: Counter, Delta: &intVal}, "42"},
		{"counter nil delta", &Metrics{ID: "counter2", MType: Counter, Delta: nil}, ""},
		{"gauge with value", &Metrics{ID: "gauge1", MType: Gauge, Value: &floatVal}, "3.14"},
		{"gauge nil value", &Metrics{ID: "gauge2", MType: Gauge, Value: nil}, ""},
		{"unsupported type", &Metrics{ID: "unknown", MType: "unknown"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewMetricString(tt.input)
			assert.Equal(t, tt.output, result)
		})
	}
}

func TestNewMetricsHTML(t *testing.T) {
	intVal1 := int64(10)
	intVal2 := int64(20)
	floatVal1 := 1.1
	floatVal2 := 2.2

	metrics := []*Metrics{
		{ID: "counter1", MType: Counter, Delta: &intVal1},
		{ID: "gauge1", MType: Gauge, Value: &floatVal1},
		{ID: "counter2", MType: Counter, Delta: &intVal2},
		{ID: "gauge2", MType: Gauge, Value: &floatVal2},
		{ID: "counter3", MType: Counter, Delta: nil}, // should skip
		{ID: "gauge3", MType: Gauge, Value: nil},     // should skip
		{ID: "unknown", MType: "unknown"},            // should skip
	}

	got := NewMetricsHTML(metrics)

	assert.Contains(t, got, "<li>counter1: 10</li>")
	assert.Contains(t, got, "<li>gauge1: 1.1</li>")
	assert.Contains(t, got, "<li>counter2: 20</li>")
	assert.Contains(t, got, "<li>gauge2: 2.2</li>")

	assert.NotContains(t, got, "<li>counter3:")
	assert.NotContains(t, got, "<li>gauge3:")
	assert.NotContains(t, got, "<li>unknown:")

	assert.Contains(t, got, "<html>")
	assert.Contains(t, got, "<ul>")
	assert.Contains(t, got, "</ul>")
	assert.Contains(t, got, "</html>")
}

func TestNewMetricFromAttributes(t *testing.T) {
	tests := []struct {
		name   string
		mType  string
		mName  string
		mValue string
		want   *Metrics
	}{
		{
			name:   "valid counter",
			mType:  Counter,
			mName:  "requests",
			mValue: "42",
			want: &Metrics{
				ID:    "requests",
				MType: Counter,
				Delta: ptrInt64(42),
			},
		},
		{
			name:   "invalid counter value",
			mType:  Counter,
			mName:  "requests",
			mValue: "notanint",
			want:   nil,
		},
		{
			name:   "valid gauge",
			mType:  Gauge,
			mName:  "temperature",
			mValue: "36.6",
			want: &Metrics{
				ID:    "temperature",
				MType: Gauge,
				Value: ptrFloat64(36.6),
			},
		},
		{
			name:   "invalid gauge value",
			mType:  Gauge,
			mName:  "temperature",
			mValue: "notafloat",
			want:   nil,
		},
		{
			name:   "unsupported metric type",
			mType:  "unknown",
			mName:  "foo",
			mValue: "123",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMetricFromAttributes(tt.mType, tt.mName, tt.mValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

// helper funcs for pointer values

func ptrInt64(v int64) *int64 {
	return &v
}

func ptrFloat64(v float64) *float64 {
	return &v
}
