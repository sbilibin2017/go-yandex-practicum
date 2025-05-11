package services

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricUpdateService_Update(t *testing.T) {
	tests := []struct {
		name  string
		setup func(mockFilterOneRepository *MockMetricUpdateFilterOneRepository, mockSaveRepository *MockMetricUpdateSaveRepository)
		want  error
	}{
		{
			name: "Error in FilterOne",
			setup: func(mockFilterOneRepository *MockMetricUpdateFilterOneRepository, mockSaveRepository *MockMetricUpdateSaveRepository) {
				mockFilterOneRepository.EXPECT().FilterOne(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
			},
			want: types.ErrInternal,
		},
		{
			name: "Error in Save",
			setup: func(mockFilterOneRepository *MockMetricUpdateFilterOneRepository, mockSaveRepository *MockMetricUpdateSaveRepository) {
				mockFilterOneRepository.EXPECT().FilterOne(gomock.Any(), gomock.Any()).Return(map[string]any{
					"type":  "counter",
					"name":  "HeapAlloc",
					"delta": int64(5),
				}, nil)
				mockSaveRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			want: types.ErrInternal,
		},
		{
			name: "Success for Gauge",
			setup: func(mockFilterOneRepository *MockMetricUpdateFilterOneRepository, mockSaveRepository *MockMetricUpdateSaveRepository) {
				mockFilterOneRepository.EXPECT().FilterOne(gomock.Any(), gomock.Any()).Return(map[string]any{
					"type":  "counter",
					"name":  "HeapAlloc",
					"delta": int64(5),
				}, nil)
				mockSaveRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: nil,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFilterOneRepository := NewMockMetricUpdateFilterOneRepository(ctrl)
	mockSaveRepository := NewMockMetricUpdateSaveRepository(ctrl)

	service := NewMetricUpdateService(mockFilterOneRepository, mockSaveRepository)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(mockFilterOneRepository, mockSaveRepository)
			metrics := []types.Metrics{
				{
					ID:    "1",
					Type:  string(types.CounterMetricType),
					Delta: 10,
				},
			}

			err := service.Update(context.Background(), metrics)
			if tt.want != nil {
				assert.Equal(t, tt.want, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricUpdateCounter(t *testing.T) {
	tests := []struct {
		name          string
		oldValue      map[string]any
		newValue      map[string]any
		expectedDelta int64
	}{
		{
			name:          "Successful Counter Update",
			oldValue:      map[string]any{"delta": int64(10)},
			newValue:      map[string]any{"delta": int64(5)},
			expectedDelta: int64(15),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := metricUpdateCounter(tt.oldValue, tt.newValue)
			assert.Equal(t, tt.expectedDelta, result["delta"], "Delta должна быть равна %d", tt.expectedDelta)
		})
	}
}

func TestMetricUpdateGauge(t *testing.T) {
	tests := []struct {
		name          string
		oldValue      map[string]any
		newValue      map[string]any
		expectedValue map[string]any
	}{
		{
			name:          "Successful Gauge Update",
			oldValue:      map[string]any{"delta": int64(10)},
			newValue:      map[string]any{"delta": int64(5)},
			expectedValue: map[string]any{"delta": int64(5)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := metricUpdateGauge(tt.oldValue, tt.newValue)
			assert.Equal(t, tt.expectedValue, result, "NewValue должна оставаться неизменной")
		})
	}
}
