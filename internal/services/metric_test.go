package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/go-yandex-practicum/internal/services"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricUpdatesService_Updates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		getter *services.MockGetter
		saver  *services.MockSaver
	}

	type args struct {
		metrics []*types.Metrics
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		want       []*types.Metrics
		wantErr    bool
		setupMocks func(fields fields, args args)
	}{
		{
			name: "success with counter metric update",
			fields: fields{
				getter: services.NewMockGetter(ctrl),
				saver:  services.NewMockSaver(ctrl),
			},
			args: args{
				metrics: []*types.Metrics{
					{
						ID:    "metric1",
						Type:  types.Counter,
						Delta: ptrInt64(10),
					},
				},
			},
			want: []*types.Metrics{
				{
					ID:    "metric1",
					Type:  types.Counter,
					Delta: ptrInt64(15), // 5 (existing) + 10 (new)
				},
			},
			setupMocks: func(f fields, a args) {
				// Simulate existing metric with Delta = 5
				f.getter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "metric1", Type: types.Counter}).
					Return(&types.Metrics{ID: "metric1", Type: types.Counter, Delta: ptrInt64(5)}, nil)

				// Expect Save with updated delta (5 + 10 = 15)
				f.saver.EXPECT().
					Save(gomock.Any(), types.Metrics{
						ID:    "metric1",
						Type:  types.Counter,
						Delta: ptrInt64(15),
					}).
					Return(nil)
			},
		},
		{
			name: "success with gauge metric, no delta update",
			fields: fields{
				getter: services.NewMockGetter(ctrl),
				saver:  services.NewMockSaver(ctrl),
			},
			args: args{
				metrics: []*types.Metrics{
					{
						ID:    "gauge1",
						Type:  types.Gauge,
						Value: ptrFloat64(3.14),
					},
				},
			},
			want: []*types.Metrics{
				{
					ID:    "gauge1",
					Type:  types.Gauge,
					Value: ptrFloat64(3.14),
				},
			},
			setupMocks: func(f fields, a args) {
				// For gauge, no Get call expected

				// Save called with input metric as is
				f.saver.EXPECT().
					Save(gomock.Any(), *a.metrics[0]).
					Return(nil)
			},
		},
		{
			name: "getter error returns error",
			fields: fields{
				getter: services.NewMockGetter(ctrl),
				saver:  services.NewMockSaver(ctrl),
			},
			args: args{
				metrics: []*types.Metrics{
					{
						ID:    "metric1",
						Type:  types.Counter,
						Delta: ptrInt64(1),
					},
				},
			},
			wantErr: true,
			setupMocks: func(f fields, a args) {
				f.getter.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db error"))
			},
		},
		{
			name: "save error returns error",
			fields: fields{
				getter: services.NewMockGetter(ctrl),
				saver:  services.NewMockSaver(ctrl),
			},
			args: args{
				metrics: []*types.Metrics{
					{
						ID:    "metric1",
						Type:  types.Gauge,
						Value: ptrFloat64(2.71),
					},
				},
			},
			wantErr: true,
			setupMocks: func(f fields, a args) {
				f.saver.EXPECT().
					Save(gomock.Any(), *a.metrics[0]).
					Return(errors.New("save failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.fields, tt.args)

			svc := services.NewMetricUpdatesService(
				services.WithMetricUpdatesGetter(tt.fields.getter),
				services.WithMetricUpdatesSaver(tt.fields.saver),
			)

			got, err := svc.Updates(context.Background(), tt.args.metrics)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, len(tt.want), len(got))
			for i := range got {
				require.Equal(t, tt.want[i].ID, got[i].ID)
				require.Equal(t, tt.want[i].Type, got[i].Type)
				if tt.want[i].Delta != nil {
					require.Equal(t, *tt.want[i].Delta, *got[i].Delta)
				} else {
					require.Nil(t, got[i].Delta)
				}
				if tt.want[i].Value != nil {
					require.InEpsilon(t, *tt.want[i].Value, *got[i].Value, 1e-9)
				} else {
					require.Nil(t, got[i].Value)
				}
			}
		})
	}
}

func TestMetricGetService_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := services.NewMockGetter(ctrl)

	svc := services.NewMetricGetService(services.WithMetricGetGetter(mockGetter))

	tests := []struct {
		name       string
		metricID   types.MetricID
		mockReturn *types.Metrics
		mockErr    error
		want       *types.Metrics
		wantErr    bool
	}{
		{
			name:       "metric found",
			metricID:   types.MetricID{ID: "id1", Type: types.Gauge},
			mockReturn: &types.Metrics{ID: "id1", Type: types.Gauge, Value: ptrFloat64(1.23)},
			mockErr:    nil,
			want:       &types.Metrics{ID: "id1", Type: types.Gauge, Value: ptrFloat64(1.23)},
			wantErr:    false,
		},
		{
			name:       "metric not found",
			metricID:   types.MetricID{ID: "id2", Type: types.Counter},
			mockReturn: nil,
			mockErr:    nil,
			want:       nil,
			wantErr:    false,
		},
		{
			name:       "getter error",
			metricID:   types.MetricID{ID: "id3", Type: types.Counter},
			mockReturn: nil,
			mockErr:    errors.New("db failure"),
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetter.EXPECT().
				Get(gomock.Any(), tt.metricID).
				Return(tt.mockReturn, tt.mockErr)

			got, err := svc.Get(context.Background(), tt.metricID)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestMetricListService_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := services.NewMockLister(ctrl)

	svc := services.NewMetricListService(services.WithMetricListLister(mockLister))

	tests := []struct {
		name       string
		mockReturn []*types.Metrics
		mockErr    error
		want       []*types.Metrics
		wantErr    bool
	}{
		{
			name: "list success",
			mockReturn: []*types.Metrics{
				{ID: "m1", Type: types.Gauge, Value: ptrFloat64(1.0)},
				{ID: "m2", Type: types.Counter, Delta: ptrInt64(100)},
			},
			mockErr: nil,
			want: []*types.Metrics{
				{ID: "m1", Type: types.Gauge, Value: ptrFloat64(1.0)},
				{ID: "m2", Type: types.Counter, Delta: ptrInt64(100)},
			},
			wantErr: false,
		},
		{
			name:       "list error",
			mockReturn: nil,
			mockErr:    errors.New("list failed"),
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLister.EXPECT().
				List(gomock.Any()).
				Return(tt.mockReturn, tt.mockErr)

			got, err := svc.List(context.Background())
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

// Helpers

func ptrInt64(i int64) *int64       { return &i }
func ptrFloat64(f float64) *float64 { return &f }
