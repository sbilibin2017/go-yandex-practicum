package services

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricListAllRepository interface {
	ListAll(ctx context.Context) ([]map[string]any, error)
}

type MetricListAllService struct {
	mlar MetricListAllRepository
}

func NewMetricListAllService(
	mlar MetricListAllRepository,
) *MetricListAllService {
	return &MetricListAllService{
		mlar: mlar,
	}
}

func (svc *MetricListAllService) ListAll(
	ctx context.Context,
) ([]types.Metrics, error) {
	metrics, err := svc.mlar.ListAll(ctx)
	if err != nil {
		return nil, types.ErrInternal
	}
	if len(metrics) == 0 {
		return nil, nil
	}
	metricsStruct, err := mapSliceToStructSlice[types.Metrics](metrics)
	if err != nil {
		return nil, err
	}
	return metricsStruct, nil
}
