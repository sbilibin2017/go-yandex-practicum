package services

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricUpdateService_Update_Error_FilterOne(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFilterOneRepository := NewMockFilterOneRepository(ctrl)
	mockSaveRepository := NewMockSaveRepository(ctrl)
	service := NewMetricUpdateService(mockFilterOneRepository, mockSaveRepository)
	metrics := []types.Metrics{
		{
			ID:    "1",
			Type:  string(types.CounterMetricType),
			Delta: 10,
		},
	}
	mockFilterOneRepository.EXPECT().FilterOne(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
	err := service.Update(context.Background(), metrics)
	assert.Equal(t, types.ErrMetricIsNotUpdated, err)
}

func TestMetricUpdateService_Update_Error_Save(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFilterOneRepository := NewMockFilterOneRepository(ctrl)
	mockSaveRepository := NewMockSaveRepository(ctrl)
	service := NewMetricUpdateService(mockFilterOneRepository, mockSaveRepository)
	metrics := []types.Metrics{
		{
			ID:    "1",
			Type:  string(types.CounterMetricType),
			Delta: 10,
		},
	}
	mockFilterOneRepository.EXPECT().FilterOne(gomock.Any(), gomock.Any()).Return(map[string]any{
		"type":  "counter",
		"name":  "HeapAlloc",
		"delta": int64(5),
	}, nil)
	mockSaveRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(assert.AnError)
	err := service.Update(context.Background(), metrics)
	assert.Equal(t, types.ErrMetricIsNotUpdated, err)
}

func TestMetricUpdateService_Update_SuccessGauge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFilterOneRepository := NewMockFilterOneRepository(ctrl)
	mockSaveRepository := NewMockSaveRepository(ctrl)
	service := NewMetricUpdateService(mockFilterOneRepository, mockSaveRepository)
	metrics := []types.Metrics{
		{
			ID:    "1",
			Type:  string(types.CounterMetricType),
			Delta: 10,
		},
	}
	mockFilterOneRepository.EXPECT().FilterOne(gomock.Any(), gomock.Any()).Return(map[string]any{
		"type":  "counter",
		"name":  "HeapAlloc",
		"delta": int64(5),
	}, nil)
	mockSaveRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
	err := service.Update(context.Background(), metrics)
	assert.NoError(t, err)
}

func TestMetricUpdateCounter(t *testing.T) {
	oldValue := map[string]any{"delta": int64(10)}
	newValue := map[string]any{"delta": int64(5)}
	result := metricUpdateCounter(oldValue, newValue)
	assert.Equal(t, int64(15), result["delta"], "Delta должна быть равна 15")
}

func TestMetricUpdateGauge(t *testing.T) {
	oldValue := map[string]any{"delta": int64(10)}
	newValue := map[string]any{"delta": int64(5)}
	result := metricUpdateGauge(oldValue, newValue)
	assert.Equal(t, newValue, result, "NewValue должна оставаться неизменной")
}
