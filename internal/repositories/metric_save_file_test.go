package repositories

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricSaveFileRepository_Save(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "metrics_test_*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	repo := NewMetricSaveFileRepository(tmpFile)
	assert.NoError(t, err)

	delta := int64(42)
	metric := types.Metrics{
		MetricID: types.MetricID{
			ID:   "requests",
			Type: types.CounterMetricType,
		},
		Delta: &delta,
	}

	err = repo.Save(context.Background(), metric)
	assert.NoError(t, err)

	var readMetric types.Metrics
	decoder := json.NewDecoder(tmpFile)
	err = decoder.Decode(&readMetric)
	assert.NoError(t, err)

	assert.Equal(t, metric.ID, readMetric.ID)
	assert.Equal(t, metric.Type, readMetric.Type)
	assert.NotNil(t, readMetric.Delta)
	assert.Equal(t, *metric.Delta, *readMetric.Delta)
}
