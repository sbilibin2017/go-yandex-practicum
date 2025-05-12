package services

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricListAllService_ListAll(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(mockRepo *MockMetricListAllRepository)
		expected []types.Metrics
		err      error
	}{
		{
			name: "Success - returns metrics list",
			setup: func(mockRepo *MockMetricListAllRepository) {
				val := float64(42.0)
				mockRepo.EXPECT().
					ListAll(context.Background()).
					Return([]types.Metrics{
						{
							MetricID: types.MetricID{ID: "load", Type: types.GaugeMetricType},
							Value:    &val,
						},
					}, nil)
			},
			expected: []types.Metrics{
				{
					MetricID: types.MetricID{ID: "load", Type: types.GaugeMetricType},
					Value:    func() *float64 { v := 42.0; return &v }(),
				},
			},
			err: nil,
		},
		{
			name: "Empty list - returns nil, nil",
			setup: func(mockRepo *MockMetricListAllRepository) {
				mockRepo.EXPECT().
					ListAll(context.Background()).
					Return([]types.Metrics{}, nil)
			},
			expected: nil,
			err:      nil,
		},
		{
			name: "Repository error - returns ErrMetricInternal",
			setup: func(mockRepo *MockMetricListAllRepository) {
				mockRepo.EXPECT().
					ListAll(context.Background()).
					Return(nil, assert.AnError)
			},
			expected: nil,
			err:      types.ErrMetricInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := NewMockMetricListAllRepository(ctrl)
			tt.setup(mockRepo)

			svc := NewMetricListAllService(mockRepo)
			result, err := svc.ListAll(context.Background())

			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
