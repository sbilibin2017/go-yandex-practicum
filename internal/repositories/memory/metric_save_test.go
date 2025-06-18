package memory

import (
	"context"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricSaveMemoryRepository_Save(t *testing.T) {
	repo := NewMetricSaveMemoryRepository()

	float64Ptr := func(f float64) *float64 {
		return &f
	}

	int64Ptr := func(i int64) *int64 {
		return &i
	}

	tests := []struct {
		name       string
		input      types.Metrics
		wantMetric types.Metrics
	}{
		{
			name: "save gauge metric",
			input: types.Metrics{
				ID:    "gauge1",
				Type:  types.Gauge,
				Value: float64Ptr(10.5),
			},
			wantMetric: types.Metrics{
				ID:    "gauge1",
				Type:  types.Gauge,
				Value: float64Ptr(10.5),
			},
		},
		{
			name: "save counter metric",
			input: types.Metrics{
				ID:    "counter1",
				Type:  types.Counter,
				Delta: int64Ptr(100),
			},
			wantMetric: types.Metrics{
				ID:    "counter1",
				Type:  types.Counter,
				Delta: int64Ptr(100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Save(context.Background(), tt.input)
			assert.NoError(t, err)

			key := types.MetricID{
				ID:   tt.input.ID,
				Type: tt.input.Type,
			}

			mu.RLock()
			got, ok := data[key]
			mu.RUnlock()

			assert.True(t, ok, "metric should be saved in memory")
			assert.Equal(t, tt.wantMetric, got)
		})
	}
}
