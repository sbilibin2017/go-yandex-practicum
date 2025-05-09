package services

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricUpdateFilterOneRepository interface {
	FilterOne(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

type MetricUpdateSaveRepository interface {
	Save(ctx context.Context, metrics types.Metrics) error
}

type MetricUpdateService struct {
	mfo MetricUpdateFilterOneRepository
	msr MetricUpdateSaveRepository
}

func NewMetricUpdateService(
	mfo MetricUpdateFilterOneRepository,
	msr MetricUpdateSaveRepository,
) *MetricUpdateService {
	return &MetricUpdateService{
		mfo: mfo,
		msr: msr,
	}
}

func (svc *MetricUpdateService) Update(
	ctx context.Context, metrics []types.Metrics,
) error {
	for _, metric := range metrics {
		currentMetric, err := svc.mfo.FilterOne(ctx, metric.MetricID)
		if err != nil {
			return types.ErrMetricIsNotUpdated
		}

		strategy := metricUpdateStrategies[metric.Type]
		updatedMetric := strategy(currentMetric, metric)

		err = svc.msr.Save(ctx, updatedMetric)
		if err != nil {
			return types.ErrMetricIsNotUpdated
		}
	}
	return nil
}

var metricUpdateStrategies = map[types.MetricType]func(
	oldValue *types.Metrics, newValue types.Metrics,
) types.Metrics{
	types.CounterMetricType: metricUpdateCounter,
	types.GaugeMetricType:   metricUpdateGauge,
}

func metricUpdateCounter(
	oldValue *types.Metrics, newValue types.Metrics,
) types.Metrics {
	if oldValue == nil {
		return newValue
	}
	*newValue.Delta = *oldValue.Delta + *newValue.Delta
	return newValue
}

func metricUpdateGauge(
	oldValue *types.Metrics, newValue types.Metrics,
) types.Metrics {
	return newValue
}
