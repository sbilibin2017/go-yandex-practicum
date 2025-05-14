package repositories

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricGetByIDFileRepository_GetByID(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "metrics_test_*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	delta1 := int64(5)
	value2 := 3.14
	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{
				ID:   "counter_metric",
				Type: types.CounterMetricType,
			},
			Delta: &delta1,
		},
		{
			MetricID: types.MetricID{
				ID:   "gauge_metric",
				Type: types.GaugeMetricType,
			},
			Value: &value2,
		},
	}
	encoder := json.NewEncoder(tmpFile)
	for _, m := range metrics {
		err := encoder.Encode(m)
		assert.NoError(t, err)
	}
	repo := NewMetricGetByIDFileRepository(tmpFile)
	targetID := types.MetricID{ID: "counter_metric", Type: types.CounterMetricType}
	result, err := repo.GetByID(context.Background(), targetID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, targetID, result.MetricID)
	assert.Equal(t, delta1, *result.Delta)
	missingID := types.MetricID{ID: "not_found", Type: types.GaugeMetricType}
	result, err = repo.GetByID(context.Background(), missingID)
	assert.NoError(t, err)
	assert.Nil(t, result)
}
