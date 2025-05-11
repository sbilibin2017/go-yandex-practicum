package services

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricGetFilterOneRepository interface {
	FilterOne(ctx context.Context, filter map[string]any) (map[string]any, error)
}

type MetricGetService struct {
	mfor MetricGetFilterOneRepository
}

func NewMetricGetService(
	mfor MetricGetFilterOneRepository,
) *MetricGetService {
	return &MetricGetService{
		mfor: mfor,
	}
}

func (svc *MetricGetService) Get(
	ctx context.Context, metricID types.MetricID,
) (*types.Metrics, error) {
	filter := structToMap(metricID)
	currentMetric, err := svc.mfor.FilterOne(ctx, filter)
	if err != nil {
		return nil, types.ErrInternal
	}
	if currentMetric == nil {
		return nil, types.ErrMetricNotFound
	}
	metric, err := mapToStruct[types.Metrics](currentMetric)
	if err != nil {
		return nil, types.ErrInternal
	}
	return &metric, nil
}
