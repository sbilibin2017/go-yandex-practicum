package repositories_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricSaveMemoryRepository_Save(t *testing.T) {
	memStorage := make(map[types.MetricID]types.Metrics)
	repo := repositories.NewMetricSaveMemoryRepository(memStorage)
	metric := types.Metrics{
		MetricID: types.MetricID{
			ID:   "test_metric",
			Type: types.CounterMetricType,
		},
		Delta: func() *int64 { v := int64(42); return &v }(),
	}
	err := repo.Save(context.Background(), metric)
	require.NoError(t, err)
	savedMetric, exists := memStorage[metric.MetricID]
	require.True(t, exists, "metric should exist in memory")
	assert.Equal(t, metric, savedMetric)
}
