package repositories

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricFileListAllRepository_ListAll(t *testing.T) {
	// Подготовка временного файла
	tmpFile, err := os.CreateTemp("", "metrics_list_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Подготовка тестовых метрик
	v1 := 3.14
	v2 := 2.71
	d1 := int64(100)
	d2 := int64(200)

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{ID: "gauge1", Type: types.GaugeMetricType},
			Value:    &v1,
		},
		{
			MetricID: types.MetricID{ID: "counter1", Type: types.CounterMetricType},
			Delta:    &d1,
		},
		{
			MetricID: types.MetricID{ID: "gauge2", Type: types.GaugeMetricType},
			Value:    &v2,
		},
		{
			MetricID: types.MetricID{ID: "counter2", Type: types.CounterMetricType},
			Delta:    &d2,
		},
	}

	// Запись метрик в файл
	encoder := json.NewEncoder(tmpFile)
	for _, m := range metrics {
		err := encoder.Encode(m)
		require.NoError(t, err)
	}

	// Открытие файла на чтение
	_, err = tmpFile.Seek(0, 0)
	require.NoError(t, err)

	repo := NewMetricFileListAllRepository(tmpFile)

	// Чтение метрик из файла
	readMetrics, err := repo.ListAll(context.Background())
	require.NoError(t, err)
	require.Len(t, readMetrics, len(metrics))

	// Сортировка для сравнения
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})
	sort.Slice(readMetrics, func(i, j int) bool {
		return readMetrics[i].ID < readMetrics[j].ID
	})

	// Проверка содержимого
	for i := range metrics {
		assert.Equal(t, metrics[i].ID, readMetrics[i].ID)
		assert.Equal(t, metrics[i].Type, readMetrics[i].Type)

		switch metrics[i].Type {
		case types.GaugeMetricType:
			require.NotNil(t, metrics[i].Value)
			require.NotNil(t, readMetrics[i].Value)
			assert.InEpsilon(t, *metrics[i].Value, *readMetrics[i].Value, 0.0001)
		case types.CounterMetricType:
			require.NotNil(t, metrics[i].Delta)
			require.NotNil(t, readMetrics[i].Delta)
			assert.Equal(t, *metrics[i].Delta, *readMetrics[i].Delta)
		}
	}
}
