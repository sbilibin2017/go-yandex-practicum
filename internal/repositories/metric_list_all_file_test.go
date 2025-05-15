package repositories

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricListAllFileRepository_ListAll(t *testing.T) {
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
		{
			MetricID: types.MetricID{
				ID:   "another_counter",
				Type: types.CounterMetricType,
			},
			Delta: &delta1,
		},
	}
	encoder := json.NewEncoder(tmpFile)
	for _, m := range metrics {
		err := encoder.Encode(m)
		assert.NoError(t, err)
	}
	repo := NewMetricListAllFileRepository(tmpFile)
	result, err := repo.ListAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "another_counter", result[0].ID)
	assert.Equal(t, "counter_metric", result[1].ID)
	assert.Equal(t, "gauge_metric", result[2].ID)
	assert.Equal(t, delta1, *result[0].Delta)
	assert.Equal(t, value2, *result[2].Value)
}
