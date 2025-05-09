package repositories

import (
	"context"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricFilterOneRepository_FilterOne(t *testing.T) {
	initialData := map[types.MetricID]types.Metrics{
		types.MetricID{ID: "1", Type: types.CounterMetricType}: {MetricID: types.MetricID{ID: "1", Type: types.CounterMetricType}, Value: nil, Delta: nil},
		types.MetricID{ID: "2", Type: types.GaugeMetricType}:   {MetricID: types.MetricID{ID: "2", Type: types.GaugeMetricType}, Value: nil, Delta: nil},
	}
	repo := NewMetricFilterOneRepository(initialData)

	t.Run("Metric exists", func(t *testing.T) {
		metricID := types.MetricID{ID: "1", Type: types.CounterMetricType}
		metric, err := repo.FilterOne(context.Background(), metricID)
		require.NoError(t, err)
		assert.NotNil(t, metric)
		assert.Equal(t, metricID, metric.MetricID)
	})

	t.Run("Metric does not exist", func(t *testing.T) {
		metricID := types.MetricID{ID: "nonexistent", Type: types.CounterMetricType}
		metric, err := repo.FilterOne(context.Background(), metricID)
		require.NoError(t, err)
		assert.Nil(t, metric)
	})
}
