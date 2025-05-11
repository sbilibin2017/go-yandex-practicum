package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricGetService_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := NewMockMetricGetFilterOneRepository(ctrl)
	metricService := NewMetricGetService(mockRepo)
	metricID := types.MetricID{ID: "metric1"}
	mockData := map[string]any{
		"id":    "metric1",
		"type":  "counter",
		"delta": 5,
	}
	tests := []struct {
		name           string
		mockReturn     map[string]any
		mockError      error
		expectedError  error
		expectedMetric *types.Metrics
	}{
		{
			name:          "successful metric retrieval",
			mockReturn:    mockData,
			mockError:     nil,
			expectedError: nil,
			expectedMetric: &types.Metrics{
				ID:    "metric1",
				Type:  "counter",
				Delta: 5,
			},
		},
		{
			name:           "metric not found",
			mockReturn:     nil,
			mockError:      nil,
			expectedError:  types.ErrMetricNotFound,
			expectedMetric: nil,
		},
		{
			name:           "error in repository",
			mockReturn:     nil,
			mockError:      errors.New("db error"),
			expectedError:  types.ErrInternal,
			expectedMetric: nil,
		},
		{
			name: "error during mapToStruct decoding",
			mockReturn: map[string]any{
				"id":    "metric1",
				"type":  "counter",
				"delta": "invalid", // некорректный тип: string вместо int64
			},
			mockError:      nil,
			expectedError:  types.ErrInternal,
			expectedMetric: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().FilterOne(gomock.Any(), gomock.Any()).Return(tt.mockReturn, tt.mockError)
			result, err := metricService.Get(context.Background(), metricID)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.expectedMetric != nil {
				assert.Equal(t, tt.expectedMetric, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
