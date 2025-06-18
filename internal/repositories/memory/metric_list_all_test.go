package memory

import (
	"context"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricListAllMemoryRepository_ListAll(t *testing.T) {
	repo := NewMetricListAllMemoryRepository()

	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	// Seed data with unordered entries
	metricsToSeed := []types.Metrics{
		{ID: "z-metric", Type: types.Gauge, Value: float64Ptr(5.0)},
		{ID: "a-metric", Type: types.Counter, Delta: int64Ptr(10)},
		{ID: "m-metric", Type: types.Gauge, Value: float64Ptr(3.14)},
	}

	// Clear existing data and seed new metrics under lock
	mu.Lock()
	data = make(map[types.MetricID]types.Metrics)
	for _, metric := range metricsToSeed {
		key := types.MetricID{ID: metric.ID, Type: metric.Type}
		data[key] = metric
	}
	mu.Unlock()

	tests := []struct {
		name      string
		want      []types.Metrics
		wantError bool
	}{
		{
			name: "list all metrics sorted by ID",
			want: []types.Metrics{
				{ID: "a-metric", Type: types.Counter, Delta: int64Ptr(10)},
				{ID: "m-metric", Type: types.Gauge, Value: float64Ptr(3.14)},
				{ID: "z-metric", Type: types.Gauge, Value: float64Ptr(5.0)},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.ListAll(context.Background())
			assert.Equal(t, tt.wantError, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}
