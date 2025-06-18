package file

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricGetByIDFileRepository_GetByID(t *testing.T) {
	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	// Helper: write multiple metrics to file (one JSON object per line)
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

	tmpFile := t.TempDir() + "/metrics.json"

	initialMetrics := []types.Metrics{
		{ID: "gauge1", Type: types.Gauge, Value: float64Ptr(12.3)},
		{ID: "counter1", Type: types.Counter, Delta: int64Ptr(100)},
	}
	writeMetricsToFile(tmpFile, initialMetrics)

	repo := NewMetricGetByIDFileRepository(tmpFile)

	tests := []struct {
		name       string
		input      types.MetricID
		wantMetric *types.Metrics
		wantErr    bool
		errMessage string
	}{
		{
			name:  "find existing gauge metric",
			input: types.MetricID{ID: "gauge1", Type: types.Gauge},
			wantMetric: &types.Metrics{
				ID:    "gauge1",
				Type:  types.Gauge,
				Value: float64Ptr(12.3),
			},
		},
		{
			name:  "find existing counter metric",
			input: types.MetricID{ID: "counter1", Type: types.Counter},
			wantMetric: &types.Metrics{
				ID:    "counter1",
				Type:  types.Counter,
				Delta: int64Ptr(100),
			},
		},
		{
			name:       "metric not found",
			input:      types.MetricID{ID: "nonexistent", Type: types.Gauge},
			wantMetric: nil,
			wantErr:    true,
			errMessage: "metric not found",
		},
		{
			name:       "file not found",
			input:      types.MetricID{ID: "any", Type: types.Gauge},
			wantMetric: nil,
			wantErr:    true,
			errMessage: "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testRepo *MetricGetByIDFileRepository
			if tt.name == "file not found" {
				// use a path that does not exist
				testRepo = NewMetricGetByIDFileRepository(tmpFile + ".notexist")
			} else {
				testRepo = repo
			}

			got, err := testRepo.GetByID(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantMetric, got)
		})
	}
}
