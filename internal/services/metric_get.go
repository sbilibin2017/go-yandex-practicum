package services

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricGetByIDRepository interface {
	GetByID(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

type MetricGetService struct {
	mfor MetricGetByIDRepository
}

func NewMetricGetService(
	mfor MetricGetByIDRepository,
) *MetricGetService {
	return &MetricGetService{
		mfor: mfor,
	}
}

func (svc *MetricGetService) Get(
	ctx context.Context, metricID types.MetricID,
) (*types.Metrics, error) {
	currentMetric, err := svc.mfor.GetByID(ctx, metricID)
	if err != nil {
		return nil, types.ErrMetricInternal
	}
	if currentMetric == nil {
		return nil, types.ErrMetricNotFound
	}
	return currentMetric, nil
}
