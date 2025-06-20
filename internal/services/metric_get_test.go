package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricGetService_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMetricGetByIDRepository(ctrl)
	svc := NewMetricGetService(mockRepo)
	ctx := context.Background()

	metricID := types.MetricID{ID: "metric1", MType: types.Counter}
	expectedMetric := &types.Metrics{
		ID:    "metric1",
		MType: types.Counter,
		Delta: nil,
		Value: nil,
	}

	t.Run("successfully get metric", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(ctx, metricID).
			Return(expectedMetric, nil).
			Times(1)

		result, err := svc.Get(ctx, metricID)
		assert.NoError(t, err)
		assert.Equal(t, expectedMetric, result)
	})

	t.Run("repository returns error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(ctx, metricID).
			Return(nil, errors.New("repo error")).
			Times(1)

		result, err := svc.Get(ctx, metricID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.EqualError(t, err, "repo error")
	})
}
