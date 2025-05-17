package repositories_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricGetByIDContextRepository_GetByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockMetricGetByIDRepository(ctrl)

	ctx := context.Background()

	metricID := types.MetricID{
		ID:   "heap_alloc",
		Type: types.GaugeMetricType,
	}

	expected := &types.Metrics{
		MetricID: metricID,
		Value:    ptrFloat64(99.9),
	}

	mockRepo.EXPECT().
		GetByID(ctx, metricID).
		Return(expected, nil)

	repo := repositories.NewMetricGetByIDContextRepository()
	repo.SetContext(mockRepo)

	result, err := repo.GetByID(ctx, metricID)

	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestMetricGetByIDContextRepository_GetByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockMetricGetByIDRepository(ctrl)

	ctx := context.Background()

	metricID := types.MetricID{
		ID:   "not_exist",
		Type: types.CounterMetricType,
	}

	mockRepo.EXPECT().
		GetByID(ctx, metricID).
		Return(nil, types.ErrMetricNotFound)

	repo := repositories.NewMetricGetByIDContextRepository()
	repo.SetContext(mockRepo)

	result, err := repo.GetByID(ctx, metricID)

	require.Nil(t, result)
	require.ErrorIs(t, err, types.ErrMetricNotFound)
}

func TestMetricGetByIDContextRepository_GetByID_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockMetricGetByIDRepository(ctrl)

	ctx := context.Background()

	metricID := types.MetricID{
		ID:   "some_metric",
		Type: types.GaugeMetricType,
	}

	expectedErr := errors.New("db connection failed")

	mockRepo.EXPECT().
		GetByID(ctx, metricID).
		Return(nil, expectedErr)

	repo := repositories.NewMetricGetByIDContextRepository()
	repo.SetContext(mockRepo)

	result, err := repo.GetByID(ctx, metricID)

	require.Nil(t, result)
	require.Equal(t, expectedErr, err)
}

func TestMetricGetByIDContextRepository_GetByID_NoStrategy(t *testing.T) {
	repo := repositories.NewMetricGetByIDContextRepository()

	ctx := context.Background()

	require.Panics(t, func() {
		_, _ = repo.GetByID(ctx, types.MetricID{
			ID:   "test",
			Type: types.CounterMetricType,
		})
	})
}
