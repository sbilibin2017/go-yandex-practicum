package services

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricUpdateCounter(t *testing.T) {
	oldDelta := int64(5)
	newDelta := int64(3)
	oldMetric := types.Metrics{
		MetricID: types.MetricID{ID: "requests", Type: types.CounterMetricType},
		Delta:    &oldDelta,
	}
	newMetric := types.Metrics{
		MetricID: types.MetricID{ID: "requests", Type: types.CounterMetricType},
		Delta:    &newDelta,
	}

	updated := metricUpdateCounter(oldMetric, newMetric)

	assert.NotNil(t, updated.Delta)
	assert.Equal(t, int64(8), *updated.Delta)
}

func TestMetricUpdateGauge(t *testing.T) {
	oldValue := float64(42.0)
	newValue := float64(99.5)
	oldMetric := types.Metrics{
		MetricID: types.MetricID{ID: "load", Type: types.GaugeMetricType},
		Value:    &oldValue,
	}
	newMetric := types.Metrics{
		MetricID: types.MetricID{ID: "load", Type: types.GaugeMetricType},
		Value:    &newValue,
	}

	updated := metricUpdateGauge(oldMetric, newMetric)

	assert.NotNil(t, updated.Value)
	assert.Equal(t, float64(99.5), *updated.Value)
}

func TestMetricUpdateService_Update(t *testing.T) {
	tests := []struct {
		name  string
		input []types.Metrics
		setup func(mockGetRepo *MockMetricUpdateGetByIDRepository, mockSaveRepo *MockMetricUpdateSaveRepository)
		want  error
	}{
		{
			name: "Counter Metric Update Success",
			input: func() []types.Metrics {
				delta := int64(3)
				return []types.Metrics{
					{
						MetricID: types.MetricID{ID: "requests", Type: types.CounterMetricType},
						Delta:    &delta,
					},
				}
			}(),
			setup: func(mockGetRepo *MockMetricUpdateGetByIDRepository, mockSaveRepo *MockMetricUpdateSaveRepository) {
				oldDelta := int64(5)
				mockGetRepo.EXPECT().
					GetByID(context.Background(), gomock.Any()).
					Return(&types.Metrics{Delta: &oldDelta}, nil)

				mockSaveRepo.EXPECT().
					Save(context.Background(), gomock.AssignableToTypeOf(types.Metrics{})).
					DoAndReturn(func(ctx context.Context, m types.Metrics) error {
						assert.Equal(t, int64(8), *m.Delta)
						return nil
					})
			},
			want: nil,
		},
		{
			name: "GetByID returns error",
			input: func() []types.Metrics {
				delta := int64(1)
				return []types.Metrics{
					{
						MetricID: types.MetricID{ID: "broken", Type: types.CounterMetricType},
						Delta:    &delta,
					},
				}
			}(),
			setup: func(mockGetRepo *MockMetricUpdateGetByIDRepository, mockSaveRepo *MockMetricUpdateSaveRepository) {
				mockGetRepo.EXPECT().
					GetByID(context.Background(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			want: types.ErrMetricInternal,
		},
		{
			name: "GetByID returns nil metric (new insert)",
			input: func() []types.Metrics {
				val := float64(42.0)
				return []types.Metrics{
					{
						MetricID: types.MetricID{ID: "new_gauge", Type: types.GaugeMetricType},
						Value:    &val,
					},
				}
			}(),
			setup: func(mockGetRepo *MockMetricUpdateGetByIDRepository, mockSaveRepo *MockMetricUpdateSaveRepository) {
				mockGetRepo.EXPECT().
					GetByID(context.Background(), gomock.Any()).
					Return(nil, nil)

				mockSaveRepo.EXPECT().
					Save(context.Background(), gomock.AssignableToTypeOf(types.Metrics{})).
					DoAndReturn(func(ctx context.Context, m types.Metrics) error {
						assert.Equal(t, 42.0, *m.Value)
						return nil
					})
			},
			want: nil,
		},
		{
			name: "Unknown metric type (no strategy)",
			input: func() []types.Metrics {
				delta := int64(1)
				return []types.Metrics{
					{
						MetricID: types.MetricID{ID: "unknown", Type: "invalid_type"},
						Delta:    &delta,
					},
				}
			}(),
			setup: func(mockGetRepo *MockMetricUpdateGetByIDRepository, mockSaveRepo *MockMetricUpdateSaveRepository) {
				mockGetRepo.EXPECT().
					GetByID(context.Background(), gomock.Any()).
					Return(&types.Metrics{}, nil)
				// Стратегия не определена → может вызвать панику, если нет проверки
			},
			want: types.ErrMetricInternal, // или panic → если нет обработки
		},
		{
			name: "Save returns error after strategy update",
			input: func() []types.Metrics {
				delta := int64(3)
				return []types.Metrics{
					{
						MetricID: types.MetricID{ID: "counter_fail", Type: types.CounterMetricType},
						Delta:    &delta,
					},
				}
			}(),
			setup: func(mockGetRepo *MockMetricUpdateGetByIDRepository, mockSaveRepo *MockMetricUpdateSaveRepository) {
				oldDelta := int64(5)
				mockGetRepo.EXPECT().
					GetByID(context.Background(), gomock.Any()).
					Return(&types.Metrics{Delta: &oldDelta}, nil)

				mockSaveRepo.EXPECT().
					Save(context.Background(), gomock.Any()).
					Return(assert.AnError)
			},
			want: types.ErrMetricInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGetRepo := NewMockMetricUpdateGetByIDRepository(ctrl)
			mockSaveRepo := NewMockMetricUpdateSaveRepository(ctrl)
			service := NewMetricUpdateService(mockGetRepo, mockSaveRepo)

			// Setup the mocks based on the current test case
			tt.setup(mockGetRepo, mockSaveRepo)

			// Act
			err := service.Update(context.Background(), tt.input)

			// Assert
			assert.ErrorIs(t, err, tt.want)
		})
	}
}
