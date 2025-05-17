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

func TestMetricListAllContextRepository_ListAll_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockMetricListAllRepository(ctrl)

	ctx := context.Background()

	expected := []types.Metrics{
		{
			MetricID: types.MetricID{
				ID:   "requests_total",
				Type: types.CounterMetricType,
			},
			Delta: ptrInt64(42),
		},
		{
			MetricID: types.MetricID{
				ID:   "heap_alloc",
				Type: types.GaugeMetricType,
			},
			Value: ptrFloat64(128.5),
		},
	}

	mockRepo.EXPECT().
		ListAll(ctx).
		Return(expected, nil)

	repo := repositories.NewMetricListAllContextRepository()
	repo.SetContext(mockRepo)

	result, err := repo.ListAll(ctx)

	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestMetricListAllContextRepository_ListAll_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockMetricListAllRepository(ctrl)

	ctx := context.Background()
	expectedErr := errors.New("db error")

	mockRepo.EXPECT().
		ListAll(ctx).
		Return(nil, expectedErr)

	repo := repositories.NewMetricListAllContextRepository()
	repo.SetContext(mockRepo)

	result, err := repo.ListAll(ctx)

	require.Nil(t, result)
	require.Equal(t, expectedErr, err)
}

func TestMetricListAllContextRepository_ListAll_NoStrategy(t *testing.T) {
	repo := repositories.NewMetricListAllContextRepository()

	ctx := context.Background()

	// Паника, если стратегия не установлена — проверка защищает от nil pointer dereference
	require.Panics(t, func() {
		_, _ = repo.ListAll(ctx)
	})
}

func ptrInt64(i int64) *int64 {
	return &i
}

func ptrFloat64(f float64) *float64 {
	return &f
}
