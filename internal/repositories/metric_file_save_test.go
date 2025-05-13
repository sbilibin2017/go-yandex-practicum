package repositories

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricFileSaveRepository_Save(t *testing.T) {
	// Создаём временный файл
	tmpFile, err := os.CreateTemp("", "metrics_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name()) // Удалим после теста
	defer tmpFile.Close()

	repo := NewMetricFileSaveRepository(tmpFile)

	metric := types.Metrics{
		MetricID: types.MetricID{
			ID:   "TestMetric",
			Type: types.GaugeMetricType,
		},
		Value: func() *float64 { v := 123.456; return &v }(),
	}

	err = repo.Save(context.Background(), metric)
	assert.NoError(t, err)

	// Закроем буфер, чтобы flush точно сработал
	err = tmpFile.Sync()
	require.NoError(t, err)

	// Считаем обратно из файла
	_, err = tmpFile.Seek(0, 0)
	require.NoError(t, err)

	var decoded types.Metrics
	decoder := json.NewDecoder(tmpFile)
	err = decoder.Decode(&decoded)
	require.NoError(t, err)

	assert.Equal(t, metric.ID, decoded.ID)
	assert.Equal(t, metric.Type, decoded.Type)
	assert.NotNil(t, decoded.Value)
	assert.InEpsilon(t, *metric.Value, *decoded.Value, 0.0001)
}
