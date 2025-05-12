package repositories_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricListAllRepository_ListAll(t *testing.T) {
	delta := func(v int64) *int64 { return &v }

	tests := []struct {
		name           string
		inputStorage   map[types.MetricID]types.Metrics
		expectedIDs    []string
		expectedDeltas []int64
	}{
		{
			name: "Multiple metrics sorted by ID",
			inputStorage: map[types.MetricID]types.Metrics{
				{ID: "bravo", Type: types.CounterMetricType}: {
					MetricID: types.MetricID{ID: "bravo", Type: types.CounterMetricType},
					Delta:    delta(20),
				},
				{ID: "alpha", Type: types.CounterMetricType}: {
					MetricID: types.MetricID{ID: "alpha", Type: types.CounterMetricType},
					Delta:    delta(10),
				},
				{ID: "charlie", Type: types.CounterMetricType}: {
					MetricID: types.MetricID{ID: "charlie", Type: types.CounterMetricType},
					Delta:    delta(30),
				},
			},
			expectedIDs:    []string{"alpha", "bravo", "charlie"},
			expectedDeltas: []int64{10, 20, 30},
		},
		{
			name:           "Empty storage returns empty slice",
			inputStorage:   map[types.MetricID]types.Metrics{},
			expectedIDs:    []string{},
			expectedDeltas: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repositories.NewMetricMemoryListAllRepository(tt.inputStorage)
			result, err := repo.ListAll(context.Background())
			require.NoError(t, err)
			require.Len(t, result, len(tt.expectedIDs))

			for i, metric := range result {
				assert.Equal(t, tt.expectedIDs[i], metric.ID, "unexpected metric ID at index %d", i)
				if metric.Delta != nil {
					assert.Equal(t, tt.expectedDeltas[i], *metric.Delta, "unexpected Delta at index %d", i)
				} else {
					assert.Nil(t, metric.Delta, "expected nil Delta at index %d", i)
				}
			}
		})
	}
}
