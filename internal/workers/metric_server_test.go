package workers

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricServerWorker_SyncStore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	memoryListRepo := NewMockMetricListAllMemoryRepository(ctrl)
	memorySaveRepo := NewMockMetricSaveMemoryRepository(ctrl)
	fileListRepo := NewMockMetricListAllFileRepository(ctrl)
	fileSaveRepo := NewMockMetricSaveFileRepository(ctrl)

	testMetrics := []types.Metrics{
		{
			MetricID: types.MetricID{ID: "foo", Type: types.GaugeMetricType},
			Value:    floatPtr(123.456),
		},
	}

	memoryListRepo.EXPECT().ListAll(gomock.Any()).Return(testMetrics, nil).Times(1)
	fileSaveRepo.EXPECT().Save(gomock.Any(), testMetrics[0]).Return(nil).Times(1)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	worker := NewMetricServerWorker(memoryListRepo, memorySaveRepo, fileListRepo, fileSaveRepo, nil, false)
	err := worker.Start(ctx)
	assert.NoError(t, err)
}

func TestMetricServerWorker_AsyncStore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	memoryListRepo := NewMockMetricListAllMemoryRepository(ctrl)
	memorySaveRepo := NewMockMetricSaveMemoryRepository(ctrl)
	fileListRepo := NewMockMetricListAllFileRepository(ctrl)
	fileSaveRepo := NewMockMetricSaveFileRepository(ctrl)

	testMetrics := []types.Metrics{
		{
			MetricID: types.MetricID{ID: "bar", Type: types.CounterMetricType},
			Delta:    intPtr(42),
		},
	}

	memoryListRepo.EXPECT().ListAll(gomock.Any()).Return(testMetrics, nil).MinTimes(1)
	fileSaveRepo.EXPECT().Save(gomock.Any(), testMetrics[0]).Return(nil).MinTimes(1)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(1200 * time.Millisecond)
		cancel()
	}()

	worker := NewMetricServerWorker(memoryListRepo, memorySaveRepo, fileListRepo, fileSaveRepo, time.NewTicker(1*time.Second), false)
	err := worker.Start(ctx)
	assert.NoError(t, err)
}

func TestMetricServerWorker_RestoreEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	memoryListRepo := NewMockMetricListAllMemoryRepository(ctrl)
	memorySaveRepo := NewMockMetricSaveMemoryRepository(ctrl)
	fileListRepo := NewMockMetricListAllFileRepository(ctrl)
	fileSaveRepo := NewMockMetricSaveFileRepository(ctrl)

	testMetrics := []types.Metrics{
		{
			MetricID: types.MetricID{ID: "baz", Type: types.GaugeMetricType},
			Value:    floatPtr(789.1),
		},
	}

	fileListRepo.EXPECT().ListAll(gomock.Any()).Return(testMetrics, nil).Times(1)
	memorySaveRepo.EXPECT().Save(gomock.Any(), testMetrics[0]).Return(nil).Times(1)

	memoryListRepo.EXPECT().ListAll(gomock.Any()).Return(testMetrics, nil).Times(1)
	fileSaveRepo.EXPECT().Save(gomock.Any(), testMetrics[0]).Return(nil).Times(1)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	worker := NewMetricServerWorker(memoryListRepo, memorySaveRepo, fileListRepo, fileSaveRepo, nil, true)
	err := worker.Start(ctx)
	assert.NoError(t, err)
}

// Helper function to create float pointers
func floatPtr(f float64) *float64 { return &f }

// Test for saveMetricsToFile
func TestSaveMetricsToFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMemoryRepo := NewMockMetricListAllMemoryRepository(ctrl)
	mockFileSaveRepo := NewMockMetricSaveFileRepository(ctrl)

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{ID: "metric1", Type: types.CounterMetricType},
			Delta:    new(int64),
		},
		{
			MetricID: types.MetricID{ID: "metric2", Type: types.GaugeMetricType},
			Value:    new(float64),
		},
	}

	// Mock ListAll and Save
	mockMemoryRepo.EXPECT().ListAll(gomock.Any()).Return(metrics, nil)
	mockFileSaveRepo.EXPECT().Save(gomock.Any(), metrics[0]).Return(nil).Times(1)
	mockFileSaveRepo.EXPECT().Save(gomock.Any(), metrics[1]).Return(nil).Times(1)

	// Call the function under test
	err := saveMetricsToFile(context.Background(), mockMemoryRepo, mockFileSaveRepo)

	// Check results
	require.NoError(t, err)
}

// Test for saveMetricsToFile with error during Save
func TestSaveMetricsToFile_ErrorSavingMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMemoryRepo := NewMockMetricListAllMemoryRepository(ctrl)
	mockFileSaveRepo := NewMockMetricSaveFileRepository(ctrl)

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{ID: "metric1", Type: types.CounterMetricType},
			Delta:    new(int64),
		},
	}

	// Mock ListAll and Save
	mockMemoryRepo.EXPECT().ListAll(gomock.Any()).Return(metrics, nil)
	mockFileSaveRepo.EXPECT().Save(gomock.Any(), metrics[0]).Return(assert.AnError)

	// Call the function under test
	err := saveMetricsToFile(context.Background(), mockMemoryRepo, mockFileSaveRepo)

	// Check that error was returned
	require.Error(t, err)
	assert.Equal(t, err, assert.AnError)
}

// Test for loadMetricsFromFile
func TestLoadMetricsFromFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileListRepo := NewMockMetricListAllFileRepository(ctrl)
	mockMemorySaveRepo := NewMockMetricSaveMemoryRepository(ctrl)

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{ID: "metric1", Type: types.CounterMetricType},
			Delta:    new(int64),
		},
		{
			MetricID: types.MetricID{ID: "metric2", Type: types.GaugeMetricType},
			Value:    new(float64),
		},
	}

	// Mock ListAll and Save
	mockFileListRepo.EXPECT().ListAll(gomock.Any()).Return(metrics, nil)
	mockMemorySaveRepo.EXPECT().Save(gomock.Any(), metrics[0]).Return(nil).Times(1)
	mockMemorySaveRepo.EXPECT().Save(gomock.Any(), metrics[1]).Return(nil).Times(1)

	// Call the function under test
	err := loadMetricsFromFile(context.Background(), mockFileListRepo, mockMemorySaveRepo)

	// Check results
	require.NoError(t, err)
}

// Test for loadMetricsFromFile with error during Save to memory
func TestLoadMetricsFromFile_ErrorSavingMetricToMemory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileListRepo := NewMockMetricListAllFileRepository(ctrl)
	mockMemorySaveRepo := NewMockMetricSaveMemoryRepository(ctrl)

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{ID: "metric1", Type: types.CounterMetricType},
			Delta:    new(int64),
		},
	}

	// Mock ListAll and Save
	mockFileListRepo.EXPECT().ListAll(gomock.Any()).Return(metrics, nil)
	mockMemorySaveRepo.EXPECT().Save(gomock.Any(), metrics[0]).Return(assert.AnError)

	// Call the function under test
	err := loadMetricsFromFile(context.Background(), mockFileListRepo, mockMemorySaveRepo)

	// Check that error was returned
	require.Error(t, err)
	assert.Equal(t, err, assert.AnError)
}

func intPtr(v int64) *int64 {
	return &v
}
