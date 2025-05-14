package services

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock repositories
	mockGetByIDRepo := NewMockMetricUpdateGetByIDRepository(ctrl)
	mockSaveRepo := NewMockMetricUpdateSaveRepository(ctrl)

	// Create the service
	service := NewMetricUpdateService(mockGetByIDRepo, mockSaveRepo)

	// Define the test cases
	tests := []struct {
		name          string
		inputMetric   types.Metrics
		mockGetByID   func()
		mockSave      func()
		expectedError error
	}{
		{
			name: "success with counter metric",
			inputMetric: types.Metrics{
				MetricID: types.MetricID{
					ID:   "1",
					Type: types.CounterMetricType,
				},
				Delta: new(int64),
			},
			mockGetByID: func() {
				mockGetByIDRepo.EXPECT().GetByID(context.Background(), types.MetricID{
					ID:   "1",
					Type: types.CounterMetricType,
				}).Return(&types.Metrics{
					MetricID: types.MetricID{
						ID:   "1",
						Type: types.CounterMetricType,
					},
					Delta: new(int64),
				}, nil)
			},
			mockSave: func() {
				mockSaveRepo.EXPECT().Save(context.Background(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "metric not found",
			inputMetric: types.Metrics{
				MetricID: types.MetricID{
					ID:   "1",
					Type: types.CounterMetricType,
				},
				Delta: new(int64),
			},
			mockGetByID: func() {
				mockGetByIDRepo.EXPECT().GetByID(context.Background(), types.MetricID{
					ID:   "1",
					Type: types.CounterMetricType,
				}).Return(nil, types.ErrMetricNotFound)
			},
			mockSave: func() {
				// no Save called
			},
			expectedError: types.ErrMetricInternal,
		},
		{
			name: "save fails",
			inputMetric: types.Metrics{
				MetricID: types.MetricID{
					ID:   "1",
					Type: types.CounterMetricType,
				},
				Delta: new(int64),
			},
			mockGetByID: func() {
				mockGetByIDRepo.EXPECT().GetByID(context.Background(), types.MetricID{
					ID:   "1",
					Type: types.CounterMetricType,
				}).Return(&types.Metrics{
					MetricID: types.MetricID{
						ID:   "1",
						Type: types.CounterMetricType,
					},
					Delta: new(int64),
				}, nil)
			},
			mockSave: func() {
				mockSaveRepo.EXPECT().Save(context.Background(), gomock.Any()).Return(types.ErrMetricInternal)
			},
			expectedError: types.ErrMetricInternal,
		},
		{
			name: "unknown metric type",
			inputMetric: types.Metrics{
				MetricID: types.MetricID{
					ID:   "2",
					Type: "unknown", // unsupported type
				},
				Delta: new(int64),
			},
			mockGetByID: func() {
				mockGetByIDRepo.EXPECT().GetByID(context.Background(), types.MetricID{
					ID:   "2",
					Type: "unknown",
				}).Return(nil, types.ErrMetricNotFound)
			},
			mockSave: func() {
				// no Save called
			},
			expectedError: types.ErrMetricInternal,
		},
	}

	// Run the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tt.mockGetByID()
			tt.mockSave()

			// Call the method
			err := service.Update(context.Background(), tt.inputMetric)

			// Assert the result
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMetricUpdateCounter(t *testing.T) {
	oldDelta := int64(10)
	newDelta := int64(20)
	oldMetric := types.Metrics{
		MetricID: types.MetricID{
			ID:   "counter1",
			Type: types.CounterMetricType,
		},
		Delta: &oldDelta,
	}
	newMetric := types.Metrics{
		MetricID: types.MetricID{
			ID:   "counter1",
			Type: types.CounterMetricType,
		},
		Delta: &newDelta,
	}

	result := metricUpdateCounter(oldMetric, newMetric)

	expectedDelta := int64(30)
	assert.Equal(t, expectedDelta, *result.Delta)
}

func TestMetricUpdateGauge(t *testing.T) {
	oldValue := 10.5
	newValue := 20.5
	oldMetric := types.Metrics{
		MetricID: types.MetricID{
			ID:   "gauge1",
			Type: types.GaugeMetricType,
		},
		Value: &oldValue,
	}
	newMetric := types.Metrics{
		MetricID: types.MetricID{
			ID:   "gauge1",
			Type: types.GaugeMetricType,
		},
		Value: &newValue,
	}

	result := metricUpdateGauge(oldMetric, newMetric)

	assert.Equal(t, newValue, *result.Value)
}
