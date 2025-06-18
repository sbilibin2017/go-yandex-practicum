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

func TestMetricListAllFileRepository_ListAll(t *testing.T) {
	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	// Helper to write metrics as JSON objects to a file (one per line)
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

	sortMetricsByID := func(metrics []types.Metrics) {
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].ID < metrics[j].ID
		})
	}

	tests := []struct {
		name          string
		metricsInFile []types.Metrics
		wantMetrics   []types.Metrics
		wantErr       bool
	}{
		{
			name:          "empty file returns empty slice",
			metricsInFile: []types.Metrics{},
			wantMetrics:   []types.Metrics{},
		},
		{
			name: "returns all unique metrics sorted by ID",
			metricsInFile: []types.Metrics{
				{ID: "z", Type: types.Gauge, Value: float64Ptr(3.14)},
				{ID: "a", Type: types.Counter, Delta: int64Ptr(7)},
				{ID: "m", Type: types.Gauge, Value: float64Ptr(2.71)},
				// duplicate id + type; should overwrite previous
				{ID: "a", Type: types.Counter, Delta: int64Ptr(42)},
			},
			wantMetrics: []types.Metrics{
				{ID: "a", Type: types.Counter, Delta: int64Ptr(42)},
				{ID: "m", Type: types.Gauge, Value: float64Ptr(2.71)},
				{ID: "z", Type: types.Gauge, Value: float64Ptr(3.14)},
			},
		},
		{
			name:    "file does not exist returns error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var repo *MetricListAllFileRepository
			if tt.name == "file does not exist returns error" {
				repo = NewMetricListAllFileRepository(t.TempDir() + "/nonexistent.json")
			} else {
				filePath := t.TempDir() + "/metrics.json"
				writeMetricsToFile(filePath, tt.metricsInFile)
				repo = NewMetricListAllFileRepository(filePath)
			}

			gotMetrics, err := repo.ListAll(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			sortMetricsByID(gotMetrics)
			sortMetricsByID(tt.wantMetrics)

			assert.Equal(t, tt.wantMetrics, gotMetrics)
		})
	}
}
