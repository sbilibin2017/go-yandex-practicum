package memory

import (
	"context"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricGetByIDMemoryRepository_GetByID(t *testing.T) {
	repo := NewMetricGetByIDMemoryRepository()

	// Prepare test data in shared memory before running tests
	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	testMetrics := []types.Metrics{
		{
			ID:    "gauge1",
			Type:  types.Gauge,
			Value: float64Ptr(3.14),
		},
		{
			ID:    "counter1",
			Type:  types.Counter,
			Delta: int64Ptr(42),
		},
	}

	// Seed the shared data map with test metrics under lock
	mu.Lock()
	for _, metric := range testMetrics {
		key := types.MetricID{ID: metric.ID, Type: metric.Type}
		data[key] = metric
	}
	mu.Unlock()

	tests := []struct {
		name      string
		inputID   types.MetricID
		want      *types.Metrics
		wantError bool
	}{
		{
			name:    "existing gauge metric",
			inputID: types.MetricID{ID: "gauge1", Type: types.Gauge},
			want: &types.Metrics{
				ID:    "gauge1",
				Type:  types.Gauge,
				Value: float64Ptr(3.14),
			},
			wantError: false,
		},
		{
			name:    "existing counter metric",
			inputID: types.MetricID{ID: "counter1", Type: types.Counter},
			want: &types.Metrics{
				ID:    "counter1",
				Type:  types.Counter,
				Delta: int64Ptr(42),
			},
			wantError: false,
		},
		{
			name:      "non-existing metric",
			inputID:   types.MetricID{ID: "notfound", Type: types.Gauge},
			want:      nil,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(context.Background(), tt.inputID)
			assert.Equal(t, tt.wantError, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}
