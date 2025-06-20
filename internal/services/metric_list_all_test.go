package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricListAllService_ListAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMetricListAllRepository(ctrl)
	svc := NewMetricListAllService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name          string
		mockSetup     func()
		expected      []*types.Metrics
		expectedError error
	}{
		{
			name: "returns list of metrics",
			mockSetup: func() {
				mockRepo.EXPECT().
					ListAll(ctx).
					Return([]*types.Metrics{
						{ID: "metric1", MType: types.Counter},
						{ID: "metric2", MType: types.Gauge},
					}, nil).
					Times(1)
			},
			expected: []*types.Metrics{
				{ID: "metric1", MType: types.Counter},
				{ID: "metric2", MType: types.Gauge},
			},
			expectedError: nil,
		},
		{
			name: "returns nil when no metrics",
			mockSetup: func() {
				mockRepo.EXPECT().
					ListAll(ctx).
					Return([]*types.Metrics{}, nil).
					Times(1)
			},
			expected:      nil,
			expectedError: nil,
		},
		{
			name: "returns error from repo",
			mockSetup: func() {
				mockRepo.EXPECT().
					ListAll(ctx).
					Return(nil, errors.New("repo error")).
					Times(1)
			},
			expected:      nil,
			expectedError: errors.New("repo error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			got, err := svc.ListAll(ctx)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
