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
		name           string
		mockSetup      func(mockRepo *MockMetricListAllRepository)
		expectedResult []types.Metrics
		isErr          bool
	}{
		{
			name: "Success - Metrics Found",
			mockSetup: func(mockRepo *MockMetricListAllRepository) {
				mockRepo.EXPECT().ListAll(gomock.Any()).Return([]map[string]any{
					{
						"id":    "1",
						"type":  "counter",
						"value": 10.0,
						"delta": int64(5),
					},
					{
						"id":    "2",
						"type":  "gauge",
						"value": 20.5,
						"delta": int64(10),
					},
				}, nil)
			},
			expectedResult: []types.Metrics{
				{
					ID:    "1",
					Type:  "counter",
					Value: 10.0,
					Delta: 5,
				},
				{
					ID:    "2",
					Type:  "gauge",
					Value: 20.5,
					Delta: 10,
				},
			},
			isErr: false,
		},
		{
			name: "Success - No Metrics Found",
			mockSetup: func(mockRepo *MockMetricListAllRepository) {
				mockRepo.EXPECT().ListAll(gomock.Any()).Return(nil, nil)
			},
			expectedResult: nil,
			isErr:          false,
		},
		{
			name: "Error - Repository Error",
			mockSetup: func(mockRepo *MockMetricListAllRepository) {
				mockRepo.EXPECT().ListAll(gomock.Any()).Return(nil, types.ErrInternal)
			},
			expectedResult: nil,
			isErr:          true,
		},
		{
			name: "Error - Failed to Map to Struct",
			mockSetup: func(mockRepo *MockMetricListAllRepository) {
				mockRepo.EXPECT().ListAll(gomock.Any()).Return([]map[string]any{
					{
						"id":    "1",
						"type":  "counter",
						"value": "invalid",
						"delta": int64(5),
					},
				}, nil)
			},
			expectedResult: nil,
			isErr:          true,
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockMetricListAllRepository(ctrl)
			tt.mockSetup(mockRepo)
			service := NewMetricListAllService(mockRepo)
			result, err := service.ListAll(context.Background())
			if tt.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
