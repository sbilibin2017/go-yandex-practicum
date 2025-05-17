package repositories_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricSaveContextRepository_Save_CounterSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockMetricSaveRepository(ctrl)

	ctx := context.Background()
	delta := int64(100)
	metric := types.Metrics{
		MetricID: types.MetricID{
			ID:   "requests_total",
			Type: types.CounterMetricType,
		},
		Delta: &delta,
	}

	mockRepo.EXPECT().
		Save(ctx, metric).
		Return(nil)

	repo := repositories.NewMetricSaveContextRepository()
	repo.SetContext(mockRepo)

	err := repo.Save(ctx, metric)
	require.NoError(t, err)
}

func TestMetricSaveContextRepository_Save_GaugeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockMetricSaveRepository(ctrl)

	ctx := context.Background()
	value := 42.42
	metric := types.Metrics{
		MetricID: types.MetricID{
			ID:   "memory_usage",
			Type: types.GaugeMetricType,
		},
		Value: &value,
	}

	mockRepo.EXPECT().
		Save(ctx, metric).
		Return(types.ErrMetricInternal)

	repo := repositories.NewMetricSaveContextRepository()
	repo.SetContext(mockRepo)

	err := repo.Save(ctx, metric)
	require.ErrorIs(t, err, types.ErrMetricInternal)
}

func TestMetricSaveContextRepository_Save_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockMetricSaveRepository(ctrl)

	ctx := context.Background()
	delta := int64(1)
	metric := types.Metrics{
		MetricID: types.MetricID{
			ID:   "unknown_metric",
			Type: types.CounterMetricType,
		},
		Delta: &delta,
	}

	mockRepo.EXPECT().
		Save(ctx, metric).
		Return(types.ErrMetricNotFound)

	repo := repositories.NewMetricSaveContextRepository()
	repo.SetContext(mockRepo)

	err := repo.Save(ctx, metric)
	require.ErrorIs(t, err, types.ErrMetricNotFound)
}
