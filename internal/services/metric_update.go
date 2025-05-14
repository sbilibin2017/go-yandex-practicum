package services

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricUpdateGetByIDRepository interface {
	GetByID(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

type MetricUpdateSaveRepository interface {
	Save(ctx context.Context, metric types.Metrics) error
}

type MetricUpdateService struct {
	mfor MetricUpdateGetByIDRepository
	msr  MetricUpdateSaveRepository
}

func NewMetricUpdateService(
	mfor MetricUpdateGetByIDRepository,
	msr MetricUpdateSaveRepository,
) *MetricUpdateService {
	return &MetricUpdateService{
		mfor: mfor,
		msr:  msr,
	}
}

func (svc *MetricUpdateService) Update(
	ctx context.Context, metric types.Metrics,
) error {
	currentMetric, err := svc.mfor.GetByID(ctx, metric.MetricID)
	if err != nil {
		return types.ErrMetricInternal
	}
	if currentMetric != nil {
		strategy, ok := metricUpdateStrategies[metric.Type]
		if !ok {
			return types.ErrMetricInternal
		}
		metric = strategy(*currentMetric, metric)
	}
	err = svc.msr.Save(ctx, metric)
	if err != nil {
		return types.ErrMetricInternal
	}
	return nil
}

var metricUpdateStrategies = map[types.MetricType]func(
	oldValue types.Metrics, newValue types.Metrics,
) types.Metrics{
	types.CounterMetricType: metricUpdateCounter,
	types.GaugeMetricType:   metricUpdateGauge,
}

func metricUpdateCounter(
	oldValue types.Metrics, newValue types.Metrics,
) types.Metrics {
	*newValue.Delta += *oldValue.Delta
	return newValue
}

func metricUpdateGauge(
	oldValue types.Metrics, newValue types.Metrics,
) types.Metrics {
	return newValue
}
