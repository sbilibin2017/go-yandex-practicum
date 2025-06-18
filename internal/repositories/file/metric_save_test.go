package file

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

func TestMetricSaveFileRepository_Save(t *testing.T) {
	// Anonymous helper functions inside test scope
	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	sortMetricsByID := func(metrics []types.Metrics) {
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].ID < metrics[j].ID
		})
	}

	writeMetricsToFile := func(path string, metrics []types.Metrics) {
		file, err := os.Create(path)
		require.NoError(t, err)
		defer file.Close()

		encoder := json.NewEncoder(file)
		for _, m := range metrics {
			err := encoder.Encode(m)
			require.NoError(t, err)
		}
	}

	readMetricsFromFile := func(path string) []types.Metrics {
		file, err := os.Open(path)
		require.NoError(t, err)
		defer file.Close()

		var metrics []types.Metrics
		decoder := json.NewDecoder(file)
		for {
			var m types.Metrics
			err := decoder.Decode(&m)
			if err != nil {
				break
			}
			metrics = append(metrics, m)
		}
		return metrics
	}

	tmpFile := t.TempDir() + "/metrics.json"

	initialMetrics := []types.Metrics{
		{ID: "gauge1", Type: types.Gauge, Value: float64Ptr(12.3)},
		{ID: "counter1", Type: types.Counter, Delta: int64Ptr(100)},
	}
	writeMetricsToFile(tmpFile, initialMetrics)

	repo := NewMetricSaveFileRepository(tmpFile)

	tests := []struct {
		name          string
		input         types.Metrics
		wantMetrics   []types.Metrics
		expectSaveErr bool
	}{
		{
			name:  "update existing gauge metric",
			input: types.Metrics{ID: "gauge1", Type: types.Gauge, Value: float64Ptr(45.6)},
			wantMetrics: []types.Metrics{
				{ID: "gauge1", Type: types.Gauge, Value: float64Ptr(45.6)},
				{ID: "counter1", Type: types.Counter, Delta: int64Ptr(100)},
			},
		},
		{
			name:  "add new metric",
			input: types.Metrics{ID: "gauge2", Type: types.Gauge, Value: float64Ptr(9.9)},
			wantMetrics: []types.Metrics{
				{ID: "gauge1", Type: types.Gauge, Value: float64Ptr(45.6)},
				{ID: "counter1", Type: types.Counter, Delta: int64Ptr(100)},
				{ID: "gauge2", Type: types.Gauge, Value: float64Ptr(9.9)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Save(context.Background(), tt.input)
			if tt.expectSaveErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			gotMetrics := readMetricsFromFile(tmpFile)

			sortMetricsByID(tt.wantMetrics)
			sortMetricsByID(gotMetrics)

			assert.Equal(t, tt.wantMetrics, gotMetrics)
		})
	}
}
