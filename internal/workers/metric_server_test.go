package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func Test_loadMetricsFromFile_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockListAllFile := NewMockMetricListAllFileRepository(ctrl)
	mockSaveMemory := NewMockMetricSaveRepository(ctrl)

	ctx := context.Background()
	metric := types.Metrics{
		MetricID: types.MetricID{ID: "test_metric", Type: types.CounterMetricType},
	}

	expectedErr := errors.New("mocked Save error")

	mockListAllFile.EXPECT().
		ListAll(ctx).
		Return([]types.Metrics{metric}, nil)

	mockSaveMemory.EXPECT().
		Save(ctx, metric).
		Return(expectedErr)

	err := loadMetricsFromFile(ctx, mockListAllFile, mockSaveMemory)
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}

func Test_saveMetricsToFile_ListAllError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockListAll := NewMockMetricListAllRepository(ctrl)
	mockSaveFile := NewMockMetricSaveFileRepository(ctrl)

	ctx := context.Background()
	expectedErr := errors.New("mocked ListAll error")

	mockListAll.EXPECT().
		ListAll(ctx).
		Return(nil, expectedErr)

	err := saveMetricsToFile(ctx, mockListAll, mockSaveFile)
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}

func Test_startMetricServerWorker_StoreInterval_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storeInterval := 1

	mockMemList := NewMockMetricListAllRepository(ctrl)
	mockFileSave := NewMockMetricSaveFileRepository(ctrl)

	expectedErr := errors.New("save failed")

	mockMemList.EXPECT().
		ListAll(gomock.Any()).
		Return([]types.Metrics{{MetricID: types.MetricID{ID: "fail", Type: types.CounterMetricType}}}, nil)

	mockFileSave.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(expectedErr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	start := time.Now()
	err := startMetricServerWorker(
		ctx,
		mockMemList,
		nil,
		nil,
		mockFileSave,
		false,
		storeInterval,
	)
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
	require.WithinDuration(t, start.Add(time.Second), time.Now(), 2*time.Second, "Should fail shortly after ticker fires")
}

func Test_startMetricServerWorker_StoreIntervalSaveCalled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storeInterval := 1

	mockMemList := NewMockMetricListAllRepository(ctrl)
	mockFileSave := NewMockMetricSaveFileRepository(ctrl)

	var delta int64 = 100
	testMetric := types.Metrics{
		MetricID: types.MetricID{ID: "test_metric", Type: types.CounterMetricType},
		Delta:    &delta,
	}

	mockMemList.EXPECT().
		ListAll(gomock.Any()).
		MinTimes(1).
		Return([]types.Metrics{testMetric}, nil)

	mockFileSave.EXPECT().
		Save(gomock.Any(), testMetric).
		MinTimes(1).
		Return(nil)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(1500 * time.Millisecond)
		cancel()
	}()

	err := startMetricServerWorker(
		ctx,
		mockMemList,
		nil,
		nil,
		mockFileSave,
		false,
		storeInterval,
	)
	require.NoError(t, err)
}

func TestStartMetricServerWorker_RestoreMetricsSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	restore := true
	storeInterval := 0

	mockFileList := NewMockMetricListAllFileRepository(ctrl)
	mockMemSave := NewMockMetricSaveRepository(ctrl)
	mockMemList := NewMockMetricListAllRepository(ctrl)
	mockFileSave := NewMockMetricSaveFileRepository(ctrl)

	var delta int64 = 10
	metric := types.Metrics{
		MetricID: types.MetricID{ID: "test_counter", Type: types.CounterMetricType},
		Delta:    &delta,
	}

	mockFileList.EXPECT().
		ListAll(gomock.Any()).
		Return([]types.Metrics{metric}, nil)

	mockMemSave.EXPECT().
		Save(gomock.Any(), metric).
		Return(nil)

	mockMemList.EXPECT().
		ListAll(gomock.Any()).
		Return([]types.Metrics{metric}, nil)

	mockFileSave.EXPECT().
		Save(gomock.Any(), metric).
		Return(nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewMetricServerWorker(
		ctx,
		mockMemList,
		mockMemSave,
		mockFileList,
		mockFileSave,
		restore,
		storeInterval,
	)(ctx)

	require.NoError(t, err)
}

func TestStartMetricServerWorker_StoreOnShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storeInterval := 0

	mockMemList := NewMockMetricListAllRepository(ctrl)
	mockFileSave := NewMockMetricSaveFileRepository(ctrl)

	var delta int64 = 42

	mockMemList.EXPECT().
		ListAll(gomock.Any()).
		Return([]types.Metrics{
			{
				MetricID: types.MetricID{ID: "shutdown_metric", Type: types.CounterMetricType},
				Delta:    &delta,
			},
		}, nil)

	mockFileSave.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(nil)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := NewMetricServerWorker(
		ctx,
		mockMemList,
		nil,
		nil,
		mockFileSave,
		false,
		storeInterval,
	)(ctx)

	require.NoError(t, err)
}

func TestStartMetricServerWorker_RestoreFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	restore := true
	storeInterval := 0

	mockFileList := NewMockMetricListAllFileRepository(ctrl)

	mockFileList.EXPECT().
		ListAll(gomock.Any()).
		Return(nil, errors.New("restore failed"))

	err := NewMetricServerWorker(
		context.Background(),
		nil,
		nil,
		mockFileList,
		nil,
		restore,
		storeInterval,
	)(context.Background())

	require.Error(t, err)
	require.Contains(t, err.Error(), "restore failed")
}

func TestStartMetricServerWorker_SaveFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storeInterval := 0

	mockMemList := NewMockMetricListAllRepository(ctrl)
	mockFileSave := NewMockMetricSaveFileRepository(ctrl)

	var delta int64 = 99

	mockMemList.EXPECT().
		ListAll(gomock.Any()).
		Return([]types.Metrics{
			{
				MetricID: types.MetricID{ID: "fail_metric", Type: types.CounterMetricType},
				Delta:    &delta,
			},
		}, nil)

	mockFileSave.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(errors.New("save failed"))

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := NewMetricServerWorker(
		ctx,
		mockMemList,
		nil,
		nil,
		mockFileSave,
		false,
		storeInterval,
	)(ctx)

	require.Error(t, err)
	require.Contains(t, err.Error(), "save failed")
}
