package repositories_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricFilterOneRepository_FilterOne(t *testing.T) {
	delta := int64(10)
	existingID := types.MetricID{
		ID:   "test_metric",
		Type: types.CounterMetricType,
	}
	existingMetric := types.Metrics{
		MetricID: existingID,
		Delta:    &delta,
	}
	storage := map[types.MetricID]types.Metrics{
		existingID: existingMetric,
	}
	repo := repositories.NewMetricGetByIDMemoryRepository(storage)

	tests := []struct {
		name          string
		inputID       types.MetricID
		expectedFound bool
		expectedValue *types.Metrics
	}{
		{
			name:          "Metric exists",
			inputID:       existingID,
			expectedFound: true,
			expectedValue: &existingMetric,
		},
		{
			name: "Metric does not exist",
			inputID: types.MetricID{
				ID:   "non_existent_metric",
				Type: types.GaugeMetricType,
			},
			expectedFound: false,
			expectedValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(context.Background(), tt.inputID)
			require.NoError(t, err)
			if tt.expectedFound {
				require.NotNil(t, result)
				assert.Equal(t, *tt.expectedValue, *result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
