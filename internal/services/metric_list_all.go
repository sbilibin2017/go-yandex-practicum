package services

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricListAllRepository interface {
	ListAll(ctx context.Context) ([]types.Metrics, error)
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
		return nil, types.ErrMetricInternal
	}
	if len(metrics) == 0 {
		return nil, nil
	}
	return metrics, nil
}
