package services

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricUpdateService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFilterOneRepo := NewMockMetricUpdateFilterOneRepository(ctrl)
	mockSaveRepo := NewMockMetricUpdateSaveRepository(ctrl)

	service := NewMetricUpdateService(mockFilterOneRepo, mockSaveRepo)

	tests := []struct {
		name          string
		inputMetrics  []types.Metrics
		setup         func()
		expectedError error
	}{
		{
			name: "Successfully update metrics",
			inputMetrics: []types.Metrics{
				{
					MetricID: types.MetricID{
						ID:   "metric1",
						Type: types.CounterMetricType,
					},
					Delta: new(int64),
				},
			},
			setup: func() {
				mockFilterOneRepo.EXPECT().
					FilterOne(gomock.Any(), gomock.Any()).
					Return(&types.Metrics{
						MetricID: types.MetricID{
							ID:   "metric1",
							Type: types.CounterMetricType,
						},
						Delta: new(int64),
					}, nil).
					Times(1)
				mockSaveRepo.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name: "Error when filtering metric",
			inputMetrics: []types.Metrics{
				{
					MetricID: types.MetricID{
						ID:   "metric1",
						Type: types.CounterMetricType,
					},
					Delta: new(int64),
				},
			},
			setup: func() {
				mockFilterOneRepo.EXPECT().
					FilterOne(gomock.Any(), gomock.Any()).
					Return(nil, types.ErrMetricIsNotUpdated).
					Times(1)
			},
			expectedError: types.ErrMetricIsNotUpdated,
		},
		{
			name: "Error when saving metric",
			inputMetrics: []types.Metrics{
				{
					MetricID: types.MetricID{
						ID:   "metric1",
						Type: types.CounterMetricType,
					},
					Delta: new(int64),
				},
			},
			setup: func() {
				mockFilterOneRepo.EXPECT().
					FilterOne(gomock.Any(), gomock.Any()).
					Return(&types.Metrics{
						MetricID: types.MetricID{
							ID:   "metric1",
							Type: types.CounterMetricType,
						},
						Delta: new(int64),
					}, nil).
					Times(1)
				mockSaveRepo.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(types.ErrMetricIsNotUpdated).
					Times(1)
			},
			expectedError: types.ErrMetricIsNotUpdated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.Update(context.Background(), tt.inputMetrics)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestMetricUpdateCounter(t *testing.T) {
	tests := []struct {
		name          string
		oldValue      *types.Metrics
		newValue      types.Metrics
		expectedValue types.Metrics
	}{
		{
			name:     "Update Counter when old value is nil",
			oldValue: nil,
			newValue: types.Metrics{
				MetricID: types.MetricID{
					ID:   "metric1",
					Type: types.CounterMetricType,
				},
				Delta: new(int64),
			},
			expectedValue: types.Metrics{
				MetricID: types.MetricID{
					ID:   "metric1",
					Type: types.CounterMetricType,
				},
				Delta: new(int64),
			},
		},
		{
			name: "Update Counter with non-nil old value",
			oldValue: &types.Metrics{
				MetricID: types.MetricID{
					ID:   "metric1",
					Type: types.CounterMetricType,
				},
				Delta: new(int64),
			},
			newValue: types.Metrics{
				MetricID: types.MetricID{
					ID:   "metric1",
					Type: types.CounterMetricType,
				},
				Delta: new(int64),
			},
			expectedValue: types.Metrics{
				MetricID: types.MetricID{
					ID:   "metric1",
					Type: types.CounterMetricType,
				},
				Delta: new(int64),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedMetric := metricUpdateCounter(tt.oldValue, tt.newValue)
			assert.Equal(t, tt.expectedValue, updatedMetric)
		})
	}
}

func TestMetricUpdateGauge(t *testing.T) {
	tests := []struct {
		name          string
		oldValue      *types.Metrics
		newValue      types.Metrics
		expectedValue types.Metrics
	}{
		{
			name:     "Update Gauge",
			oldValue: nil,
			newValue: types.Metrics{
				MetricID: types.MetricID{
					ID:   "metric1",
					Type: types.GaugeMetricType,
				},
				Value: new(float64),
			},
			expectedValue: types.Metrics{
				MetricID: types.MetricID{
					ID:   "metric1",
					Type: types.GaugeMetricType,
				},
				Value: new(float64),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedMetric := metricUpdateGauge(tt.oldValue, tt.newValue)
			assert.Equal(t, tt.expectedValue, updatedMetric)
		})
	}
}
