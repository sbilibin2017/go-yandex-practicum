package services

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ptrInt64 := func(v int64) *int64 {
		return &v
	}

	mockGetByIDRepo := NewMockMetricUpdateGetByIDRepository(ctrl)
	mockSaveRepo := NewMockMetricUpdateSaveRepository(ctrl)

	service := NewMetricUpdatesService(mockGetByIDRepo, mockSaveRepo)

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
				Delta: ptrInt64(5),
			},
			mockGetByID: func() {
				mockGetByIDRepo.EXPECT().GetByID(gomock.Any(), types.MetricID{
					ID:   "1",
					Type: types.CounterMetricType,
				}).Return(&types.Metrics{
					MetricID: types.MetricID{
						ID:   "1",
						Type: types.CounterMetricType,
					},
					Delta: ptrInt64(3),
				}, nil)
			},
			mockSave: func() {
				mockSaveRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
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
				Delta: ptrInt64(1),
			},
			mockGetByID: func() {
				mockGetByIDRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(nil, types.ErrMetricNotFound)
			},
			expectedError: types.ErrMetricInternal,
		},
		{
			name: "GetByID returns internal error",
			inputMetric: types.Metrics{
				MetricID: types.MetricID{
					ID:   "1",
					Type: types.CounterMetricType,
				},
				Delta: ptrInt64(1),
			},
			mockGetByID: func() {
				mockGetByIDRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(nil, types.ErrMetricInternal)
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
				Delta: ptrInt64(2),
			},
			mockGetByID: func() {
				mockGetByIDRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&types.Metrics{
					MetricID: types.MetricID{
						ID:   "1",
						Type: types.CounterMetricType,
					},
					Delta: ptrInt64(3),
				}, nil)
			},
			mockSave: func() {
				mockSaveRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(types.ErrMetricInternal)
			},
			expectedError: types.ErrMetricInternal,
		},
		{
			name: "unknown metric type",
			inputMetric: types.Metrics{
				MetricID: types.MetricID{
					ID:   "unknown_id",
					Type: "foobar",
				},
			},
			expectedError: types.ErrMetricInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockGetByID != nil {
				tt.mockGetByID()
			}
			if tt.mockSave != nil {
				tt.mockSave()
			}

			_, err := service.Updates(context.Background(), []types.Metrics{tt.inputMetric})

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAggregateMetrics_CounterAggregation(t *testing.T) {
	id := "requests_total"
	d1 := int64(10)
	d2 := int64(20)

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{ID: id, Type: types.CounterMetricType},
			Delta:    &d1,
		},
		{
			MetricID: types.MetricID{ID: id, Type: types.CounterMetricType},
			Delta:    &d2,
		},
	}

	result, err := aggregateMetrics(metrics)
	require.NoError(t, err)

	assert.Len(t, result, 1)
	assert.Equal(t, id, result[0].ID)
	assert.NotNil(t, result[0].Delta)
	assert.Equal(t, int64(30), *result[0].Delta)
}

func TestAggregateMetrics_GaugePreserved(t *testing.T) {
	id := "memory_usage"
	v := 42.5

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{ID: id, Type: types.GaugeMetricType},
			Value:    &v,
		},
	}

	result, err := aggregateMetrics(metrics)
	require.NoError(t, err)

	assert.Len(t, result, 1)
	assert.Equal(t, id, result[0].ID)
	assert.NotNil(t, result[0].Value)
	assert.Equal(t, v, *result[0].Value)
}

func TestAggregateMetrics_MultipleMetricIDs(t *testing.T) {
	id1 := "counter1"
	id2 := "gauge1"
	delta := int64(5)
	value := 3.14

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{ID: id1, Type: types.CounterMetricType},
			Delta:    &delta,
		},
		{
			MetricID: types.MetricID{ID: id2, Type: types.GaugeMetricType},
			Value:    &value,
		},
	}

	result, err := aggregateMetrics(metrics)
	require.NoError(t, err)

	assert.Len(t, result, 2)
}

func TestAggregateMetrics_NilValuesIgnored(t *testing.T) {
	id := "test_counter"
	delta := int64(10)

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{ID: id, Type: types.CounterMetricType},
			Delta:    &delta,
		},
		{
			MetricID: types.MetricID{ID: id, Type: types.CounterMetricType},
			Delta:    nil,
		},
	}

	result, err := aggregateMetrics(metrics)
	require.NoError(t, err)

	assert.Len(t, result, 1)
	assert.Equal(t, int64(10), *result[0].Delta)
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
