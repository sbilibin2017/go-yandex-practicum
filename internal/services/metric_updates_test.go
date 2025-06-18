package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricUpdatesService_Updates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetRepo := NewMockMetricUpdateGetByIDRepository(ctrl)
	mockSaveRepo := NewMockMetricUpdateSaveRepository(ctrl)

	svc := NewMetricUpdatesService(mockGetRepo, mockSaveRepo)
	ctx := context.Background()

	tests := []struct {
		name            string
		inputMetrics    []*types.Metrics
		mockGetSetup    func()
		mockSaveSetup   func()
		expectedMetrics []*types.Metrics
		expectedErr     error
	}{
		{
			name: "successful update counter metric with existing delta",
			inputMetrics: []*types.Metrics{
				{ID: "counter1", Type: types.Counter, Delta: func() *int64 { v := int64(5); return &v }()},
			},
			mockGetSetup: func() {
				mockGetRepo.EXPECT().
					GetByID(ctx, types.MetricID{ID: "counter1", Type: types.Counter}).
					Return(&types.Metrics{ID: "counter1", Type: types.Counter, Delta: func() *int64 { v := int64(10); return &v }()}, nil)
			},
			mockSaveSetup: func() {
				mockSaveRepo.EXPECT().
					Save(ctx, gomock.AssignableToTypeOf(types.Metrics{})).
					DoAndReturn(func(_ context.Context, m types.Metrics) error {
						if m.ID != "counter1" || m.Type != types.Counter {
							return errors.New("wrong metric saved")
						}
						if m.Delta == nil || *m.Delta != 15 {
							return errors.New("wrong delta value")
						}
						return nil
					})
			},
			expectedMetrics: []*types.Metrics{
				{ID: "counter1", Type: types.Counter, Delta: func() *int64 { v := int64(15); return &v }()},
			},
			expectedErr: nil,
		},
		{
			name: "metric without existing delta",
			inputMetrics: []*types.Metrics{
				{ID: "counter2", Type: types.Counter, Delta: nil},
			},
			mockGetSetup: func() {
				mockGetRepo.EXPECT().
					GetByID(ctx, types.MetricID{ID: "counter2", Type: types.Counter}).
					Return(nil, nil)
			},
			mockSaveSetup: func() {
				mockSaveRepo.EXPECT().
					Save(ctx, gomock.AssignableToTypeOf(types.Metrics{})).
					DoAndReturn(func(_ context.Context, m types.Metrics) error {
						if m.ID != "counter2" || m.Type != types.Counter {
							return errors.New("wrong metric saved")
						}
						if m.Delta == nil || *m.Delta != 0 {
							return errors.New("wrong delta value")
						}
						return nil
					})
			},
			expectedMetrics: []*types.Metrics{
				{ID: "counter2", Type: types.Counter, Delta: func() *int64 { v := int64(0); return &v }()},
			},
			expectedErr: nil,
		},
		{
			name: "get by ID repository error",
			inputMetrics: []*types.Metrics{
				{ID: "counter3", Type: types.Counter, Delta: func() *int64 { v := int64(1); return &v }()},
			},
			mockGetSetup: func() {
				mockGetRepo.EXPECT().
					GetByID(ctx, types.MetricID{ID: "counter3", Type: types.Counter}).
					Return(nil, errors.New("db error"))
			},
			mockSaveSetup:   func() {},
			expectedMetrics: nil,
			expectedErr:     errors.New("db error"),
		},
		{
			name: "save repository error",
			inputMetrics: []*types.Metrics{
				{ID: "counter4", Type: types.Counter, Delta: func() *int64 { v := int64(2); return &v }()},
			},
			mockGetSetup: func() {
				mockGetRepo.EXPECT().
					GetByID(ctx, types.MetricID{ID: "counter4", Type: types.Counter}).
					Return(&types.Metrics{ID: "counter4", Type: types.Counter, Delta: func() *int64 { v := int64(3); return &v }()}, nil)
			},
			mockSaveSetup: func() {
				mockSaveRepo.EXPECT().
					Save(ctx, gomock.AssignableToTypeOf(types.Metrics{})).
					Return(errors.New("save error"))
			},
			expectedMetrics: nil,
			expectedErr:     errors.New("save error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockGetSetup()
			tt.mockSaveSetup()

			got, err := svc.Updates(ctx, tt.inputMetrics)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, len(tt.expectedMetrics))
				for i, want := range tt.expectedMetrics {
					gotMetric := got[i]
					assert.Equal(t, want.ID, gotMetric.ID)
					assert.Equal(t, want.Type, gotMetric.Type)
					if want.Delta == nil {
						assert.Nil(t, gotMetric.Delta)
					} else {
						assert.NotNil(t, gotMetric.Delta)
						assert.Equal(t, *want.Delta, *gotMetric.Delta)
					}
				}
			}
		})
	}
}
