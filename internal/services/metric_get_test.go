package services

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricGetService_Get(t *testing.T) {
	tests := []struct {
		name     string
		input    types.MetricID
		setup    func(mockRepo *MockMetricGetByIDRepository)
		expected *types.Metrics
		err      error
	}{
		{
			name:  "Success - Metric Found",
			input: types.MetricID{ID: "load", Type: types.GaugeMetricType},
			setup: func(mockRepo *MockMetricGetByIDRepository) {
				val := float64(42.0)
				mockRepo.EXPECT().
					GetByID(context.Background(), gomock.Any()).
					Return(&types.Metrics{
						MetricID: types.MetricID{ID: "load", Type: types.GaugeMetricType},
						Value:    &val,
					}, nil)
			},
			expected: &types.Metrics{
				MetricID: types.MetricID{ID: "load", Type: types.GaugeMetricType},
				Value:    func() *float64 { v := 42.0; return &v }(),
			},
			err: nil,
		},
		{
			name:  "Error - Repository returns error",
			input: types.MetricID{ID: "broken", Type: types.CounterMetricType},
			setup: func(mockRepo *MockMetricGetByIDRepository) {
				mockRepo.EXPECT().
					GetByID(context.Background(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			expected: nil,
			err:      types.ErrMetricInternal,
		},
		{
			name:  "Not Found - Repository returns nil metric",
			input: types.MetricID{ID: "missing", Type: types.CounterMetricType},
			setup: func(mockRepo *MockMetricGetByIDRepository) {
				mockRepo.EXPECT().
					GetByID(context.Background(), gomock.Any()).
					Return(nil, nil)
			},
			expected: nil,
			err:      types.ErrMetricNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := NewMockMetricGetByIDRepository(ctrl)
			tt.setup(mockRepo)

			svc := NewMetricGetService(mockRepo)

			result, err := svc.Get(context.Background(), tt.input)

			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
