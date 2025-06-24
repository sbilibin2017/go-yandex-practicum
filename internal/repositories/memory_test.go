package repositories

import (
	"context"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

// clearMemory resets the in-memory storage before tests.
func clearMemory() {
	muMemory.Lock()
	defer muMemory.Unlock()
	for k := range data {
		delete(data, k)
	}
}

// helpers to get pointer of values easily
func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func TestMetricMemorySaveRepository_Save(t *testing.T) {
	clearMemory()
	ctx := context.Background()
	repo := NewMetricMemorySaveRepository()

	tests := []struct {
		name   string
		metric types.Metrics
	}{
		{
			"Save gauge metric",
			types.Metrics{
				ID:    "id1",
				Type:  types.Gauge,
				Value: float64Ptr(1.1),
				Delta: nil,
			},
		},
		{
			"Save counter metric",
			types.Metrics{
				ID:    "id2",
				Type:  types.Counter,
				Value: nil,
				Delta: int64Ptr(10),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Save(ctx, tt.metric)
			assert.NoError(t, err)

			muMemory.RLock()
			storedMetric, ok := data[types.MetricID{ID: tt.metric.ID, Type: tt.metric.Type}]
			muMemory.RUnlock()

			assert.True(t, ok)
			assert.Equal(t, tt.metric, storedMetric)
		})
	}
}

func TestMetricMemoryGetRepository_Get(t *testing.T) {
	clearMemory()
	ctx := context.Background()
	saveRepo := NewMetricMemorySaveRepository()
	getRepo := NewMetricMemoryGetRepository()

	preMetrics := []types.Metrics{
		{
			ID:    "id1",
			Type:  types.Gauge,
			Value: float64Ptr(1.1),
			Delta: nil,
		},
		{
			ID:    "id2",
			Type:  types.Counter,
			Value: nil,
			Delta: int64Ptr(20),
		},
	}

	for _, m := range preMetrics {
		_ = saveRepo.Save(ctx, m)
	}

	tests := []struct {
		name       string
		metricID   types.MetricID
		wantMetric *types.Metrics
	}{
		{
			"Get existing gauge",
			types.MetricID{ID: "id1", Type: types.Gauge},
			&preMetrics[0],
		},
		{
			"Get existing counter",
			types.MetricID{ID: "id2", Type: types.Counter},
			&preMetrics[1],
		},
		{
			"Get non-existent metric",
			types.MetricID{ID: "id3", Type: types.Gauge},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRepo.Get(ctx, tt.metricID)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantMetric, got)
		})
	}
}

func TestMetricMemoryListRepository_List(t *testing.T) {
	clearMemory()
	ctx := context.Background()
	saveRepo := NewMetricMemorySaveRepository()
	listRepo := NewMetricMemoryListRepository()

	metrics := []types.Metrics{
		{
			ID:    "b",
			Type:  types.Gauge,
			Value: float64Ptr(3.3),
			Delta: nil,
		},
		{
			ID:    "a",
			Type:  types.Counter,
			Value: nil,
			Delta: int64Ptr(15),
		},
		{
			ID:    "c",
			Type:  types.Gauge,
			Value: float64Ptr(7.7),
			Delta: nil,
		},
	}

	for _, m := range metrics {
		_ = saveRepo.Save(ctx, m)
	}

	t.Run("List all metrics sorted by ID", func(t *testing.T) {
		list, err := listRepo.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, list, len(metrics))

		expectedOrder := []string{"a", "b", "c"}
		for i, metric := range list {
			assert.Equal(t, expectedOrder[i], metric.ID)
		}
	})
}
