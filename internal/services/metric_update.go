package services

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricUpdateFilterOneRepository interface {
	FilterOne(ctx context.Context, filter map[string]any) (map[string]any, error)
}

type MetricUpdateSaveRepository interface {
	Save(ctx context.Context, data map[string]any) error
}

type MetricUpdateService struct {
	mfor MetricUpdateFilterOneRepository
	msr  MetricUpdateSaveRepository
}

func NewMetricUpdateService(
	mfor MetricUpdateFilterOneRepository,
	msr MetricUpdateSaveRepository,
) *MetricUpdateService {
	return &MetricUpdateService{
		mfor: mfor,
		msr:  msr,
	}
}

func (svc *MetricUpdateService) Update(
	ctx context.Context, metrics []types.Metrics,
) error {
	for _, metric := range metrics {
		newMetric := structToMap(metric)

		currentMetric, err := svc.mfor.FilterOne(ctx, newMetric)
		if err != nil {
			return types.ErrInternal
		}
		if currentMetric != nil {
			strategy := metricUpdateStrategies[metric.Type]
			newMetric = strategy(currentMetric, newMetric)
		}

		err = svc.msr.Save(ctx, newMetric)
		if err != nil {
			return types.ErrInternal
		}
	}
	return nil
}

var metricUpdateStrategies = map[string]func(
	oldValue map[string]any, newValue map[string]any,
) map[string]any{
	types.CounterMetricType: metricUpdateCounter,
	types.GaugeMetricType:   metricUpdateGauge,
}

func metricUpdateCounter(
	oldValue map[string]any, newValue map[string]any,
) map[string]any {
	newValue["delta"] = newValue["delta"].(int64) + oldValue["delta"].(int64)
	return newValue
}

func metricUpdateGauge(
	oldValue map[string]any, newValue map[string]any,
) map[string]any {
	return newValue
}
