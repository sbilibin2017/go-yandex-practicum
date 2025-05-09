package repositories

import (
	"context"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricMemorySaveRepository_Save(t *testing.T) {
	initialData := map[types.MetricID]types.Metrics{}
	repo := NewMetricMemorySaveRepository(initialData)
	metric := types.Metrics{
		MetricID: types.MetricID{
			ID:   "1",
			Type: types.CounterMetricType,
		},
		Delta: nil,
		Value: nil,
	}
	err := repo.Save(context.Background(), metric)
	require.NoError(t, err)
	assert.Contains(t, repo.data, metric.MetricID)
	assert.Equal(t, metric, repo.data[metric.MetricID])
}
