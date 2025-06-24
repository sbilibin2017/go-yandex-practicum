package repositories

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func createTempFile(t *testing.T) (string, func()) {
	f, err := os.CreateTemp("", "metrics_test_*.json")
	require.NoError(t, err)
	f.Close()
	return f.Name(), func() { os.Remove(f.Name()) }
}

func TestMetricFileSaveRepository_Save(t *testing.T) {
	tmpFile, cleanup := createTempFile(t)
	defer cleanup()

	repo := NewMetricFileSaveRepository(WithMetricFileSaveRepositoryPath(tmpFile))
	ctx := context.Background()

	metric := types.Metrics{
		ID:    "metric1",
		Type:  types.Gauge,
		Value: float64Ptr(123.45),
		Delta: nil,
	}

	err := repo.Save(ctx, metric)
	require.NoError(t, err)

	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	var savedMetric types.Metrics
	err = json.Unmarshal(data, &savedMetric)
	require.NoError(t, err)

	assert.Equal(t, metric.ID, savedMetric.ID)
	assert.Equal(t, metric.Type, savedMetric.Type)
	assert.NotNil(t, savedMetric.Value)
	assert.Equal(t, *metric.Value, *savedMetric.Value)
	assert.Nil(t, savedMetric.Delta)
}

func TestMetricFileGetRepository_Get(t *testing.T) {
	tmpFile, cleanup := createTempFile(t)
	defer cleanup()

	ctx := context.Background()

	metrics := []types.Metrics{
		{ID: "m1", Type: types.Counter, Value: nil, Delta: int64Ptr(10)},
		{ID: "m2", Type: types.Gauge, Value: float64Ptr(3.14), Delta: nil},
	}

	f, err := os.OpenFile(tmpFile, os.O_WRONLY, 0644)
	require.NoError(t, err)
	for _, m := range metrics {
		b, err := json.Marshal(m)
		require.NoError(t, err)
		_, err = f.Write(append(b, '\n'))
		require.NoError(t, err)
	}
	f.Close()

	repo := NewMetricFileGetRepository(WithMetricFileGetRepositoryPath(tmpFile))

	tests := []struct {
		name       string
		id         types.MetricID
		wantMetric *types.Metrics
		wantErr    bool
	}{
		{
			name:       "existing counter metric",
			id:         types.MetricID{ID: "m1", Type: types.Counter},
			wantMetric: &metrics[0],
			wantErr:    false,
		},
		{
			name:       "existing gauge metric",
			id:         types.MetricID{ID: "m2", Type: types.Gauge},
			wantMetric: &metrics[1],
			wantErr:    false,
		},
		{
			name:       "non-existing metric",
			id:         types.MetricID{ID: "unknown", Type: types.Gauge},
			wantMetric: nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMetric, err := repo.Get(ctx, tt.id)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantMetric, gotMetric)
			}
		})
	}
}

func TestMetricFileListRepository_List(t *testing.T) {
	tmpFile, cleanup := createTempFile(t)
	defer cleanup()

	ctx := context.Background()

	metrics := []types.Metrics{
		{ID: "a", Type: types.Counter, Value: nil, Delta: int64Ptr(1)},
		{ID: "b", Type: types.Gauge, Value: float64Ptr(2.0), Delta: nil},
		{ID: "c", Type: types.Counter, Value: nil, Delta: int64Ptr(3)},
	}

	f, err := os.OpenFile(tmpFile, os.O_WRONLY, 0644)
	require.NoError(t, err)
	for _, m := range metrics {
		b, err := json.Marshal(m)
		require.NoError(t, err)
		_, err = f.Write(append(b, '\n'))
		require.NoError(t, err)
	}
	f.Close()

	repo := NewMetricFileListRepository(WithMetricFileListRepositoryPath(tmpFile))

	t.Run("list all metrics", func(t *testing.T) {
		list, err := repo.List(ctx)
		require.NoError(t, err)

		expectedOrder := []string{"a", "b", "c"}
		require.Len(t, list, len(metrics))
		for i, m := range list {
			assert.Equal(t, expectedOrder[i], m.ID)
		}
	})
}
