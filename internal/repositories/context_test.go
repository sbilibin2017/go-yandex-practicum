package repositories_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricContextSaveRepository_Save(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	metric := types.Metrics{ID: "id1", MType: "gauge"}

	tests := []struct {
		name      string
		mockSetup func(m *repositories.MockMetricSaver)
		wantErr   bool
	}{
		{
			name: "success save",
			mockSetup: func(m *repositories.MockMetricSaver) {
				m.EXPECT().Save(ctx, metric).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error on save",
			mockSetup: func(m *repositories.MockMetricSaver) {
				m.EXPECT().Save(ctx, metric).Return(errors.New("save error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSaver := repositories.NewMockMetricSaver(ctrl)
			repo := repositories.NewMetricContextSaveRepository()
			repo.SetContext(mockSaver)

			tt.mockSetup(mockSaver)

			err := repo.Save(ctx, metric)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricContextGetRepository_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	id := types.MetricID{ID: "id1", MType: "counter"}
	expectedMetric := &types.Metrics{ID: "id1", MType: "counter"}

	tests := []struct {
		name       string
		mockSetup  func(m *repositories.MockMetricGetter)
		wantMetric *types.Metrics
		wantErr    bool
	}{
		{
			name: "success get",
			mockSetup: func(m *repositories.MockMetricGetter) {
				m.EXPECT().Get(ctx, id).Return(expectedMetric, nil)
			},
			wantMetric: expectedMetric,
			wantErr:    false,
		},
		{
			name: "not found metric",
			mockSetup: func(m *repositories.MockMetricGetter) {
				m.EXPECT().Get(ctx, id).Return(nil, nil)
			},
			wantMetric: nil,
			wantErr:    false,
		},
		{
			name: "error on get",
			mockSetup: func(m *repositories.MockMetricGetter) {
				m.EXPECT().Get(ctx, id).Return(nil, errors.New("get error"))
			},
			wantMetric: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetter := repositories.NewMockMetricGetter(ctrl)
			repo := repositories.NewMetricContextGetRepository()
			repo.SetContext(mockGetter)

			tt.mockSetup(mockGetter)

			metric, err := repo.Get(ctx, id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, metric)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMetric, metric)
			}
		})
	}
}

func TestMetricContextListRepository_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	metricsList := []*types.Metrics{
		{ID: "id1", MType: "gauge"},
		{ID: "id2", MType: "counter"},
	}

	tests := []struct {
		name      string
		mockSetup func(m *repositories.MockMetricLister)
		wantList  []*types.Metrics
		wantErr   bool
	}{
		{
			name: "success list",
			mockSetup: func(m *repositories.MockMetricLister) {
				m.EXPECT().List(ctx).Return(metricsList, nil)
			},
			wantList: metricsList,
			wantErr:  false,
		},
		{
			name: "error on list",
			mockSetup: func(m *repositories.MockMetricLister) {
				m.EXPECT().List(ctx).Return(nil, errors.New("list error"))
			},
			wantList: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLister := repositories.NewMockMetricLister(ctrl)
			repo := repositories.NewMetricContextListRepository()
			repo.SetContext(mockLister)

			tt.mockSetup(mockLister)

			list, err := repo.List(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, list)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantList, list)
			}
		})
	}
}
