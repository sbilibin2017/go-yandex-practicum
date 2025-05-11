package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateMetricKey(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		expected string
	}{
		{
			name: "Test with id and type",
			data: map[string]any{
				"id":   "metric1",
				"type": "counter",
			},
			expected: "metric1:counter",
		},
		{
			name: "Test with different id and type",
			data: map[string]any{
				"id":   "metric2",
				"type": "gauge",
			},
			expected: "metric2:gauge",
		},
		{
			name: "Test with empty id",
			data: map[string]any{
				"id":   "",
				"type": "counter",
			},
			expected: ":counter",
		},
		{
			name: "Test with empty type",
			data: map[string]any{
				"id":   "metric1",
				"type": "",
			},
			expected: "metric1:",
		},
		{
			name: "Test with both empty id and type",
			data: map[string]any{
				"id":   "",
				"type": "",
			},
			expected: ":",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateMetricKey(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}
