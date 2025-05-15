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

type MetricUpdatesService struct {
	mfor MetricUpdateGetByIDRepository
	msr  MetricUpdateSaveRepository
}

func NewMetricUpdatesService(
	mfor MetricUpdateGetByIDRepository,
	msr MetricUpdateSaveRepository,
) *MetricUpdatesService {
	return &MetricUpdatesService{
		mfor: mfor,
		msr:  msr,
	}
}

func (svc *MetricUpdatesService) Updates(
	ctx context.Context, metrics []types.Metrics,
) ([]types.Metrics, error) {
	metrics, err := aggregateMetrics(metrics)
	if err != nil {
		return nil, types.ErrMetricInternal
	}
	for _, metric := range metrics {
		strategy := metricUpdateStrategies[metric.Type]
		currentMetric, err := svc.mfor.GetByID(ctx, metric.MetricID)
		if err != nil {
			return nil, types.ErrMetricInternal
		}
		if currentMetric != nil {
			metric = strategy(*currentMetric, metric)
		}
		if err := svc.msr.Save(ctx, metric); err != nil {
			return nil, types.ErrMetricInternal
		}
	}
	return metrics, nil
}

func aggregateMetrics(metrics []types.Metrics) ([]types.Metrics, error) {
	metrcsMap := make(map[types.MetricID]types.Metrics)
	for _, m := range metrics {
		_, ok := metricUpdateStrategies[m.Type]
		if !ok {
			return nil, types.ErrMetricInternal
		}
		if existing, ok := metrcsMap[m.MetricID]; ok {
			switch m.Type {
			case types.CounterMetricType:
				if existing.Delta != nil && m.Delta != nil {
					sum := *existing.Delta + *m.Delta
					existing.Delta = &sum
				}
			case types.GaugeMetricType:
				existing.Value = m.Value
			}
			metrcsMap[m.MetricID] = existing
		} else {
			metrcsMap[m.MetricID] = m
		}
	}
	var metricsAggregated []types.Metrics
	for _, m := range metrcsMap {
		metricsAggregated = append(metricsAggregated, m)
	}
	return metricsAggregated, nil
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
