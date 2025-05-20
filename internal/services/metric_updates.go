package services

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricUpdateGetByIDRepository описывает интерфейс репозитория для получения метрики по ID.
type MetricUpdateGetByIDRepository interface {
	// GetByID возвращает метрику по заданному ID.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - id: идентификатор метрики.
	//
	// Возвращает:
	//   - указатель на метрику, если она найдена.
	//   - ошибку в случае проблемы с запросом.
	GetByID(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

// MetricUpdateSaveRepository описывает интерфейс репозитория для сохранения метрики.
type MetricUpdateSaveRepository interface {
	// Save сохраняет метрику.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - metric: метрика для сохранения.
	//
	// Возвращает ошибку в случае неудачи.
	Save(ctx context.Context, metric types.Metrics) error
}

// MetricUpdatesService реализует логику обновления метрик.
type MetricUpdatesService struct {
	mfor MetricUpdateGetByIDRepository
	msr  MetricUpdateSaveRepository
}

// NewMetricUpdatesService создает новый сервис обновления метрик.
//
// Параметры:
//   - mfor: репозиторий для получения метрики по ID.
//   - msr: репозиторий для сохранения метрики.
//
// Возвращает:
//   - новый экземпляр MetricUpdatesService.
func NewMetricUpdatesService(
	mfor MetricUpdateGetByIDRepository,
	msr MetricUpdateSaveRepository,
) *MetricUpdatesService {
	return &MetricUpdatesService{
		mfor: mfor,
		msr:  msr,
	}
}

// Updates обновляет список метрик.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - metrics: список метрик для обновления.
//
// Логика:
//   - Агрегирует метрики по ID.
//   - Для каждой метрики применяет стратегию обновления (в зависимости от типа).
//   - Получает текущую метрику из репозитория.
//   - Обновляет метрику согласно стратегии.
//   - Сохраняет обновленную метрику.
//
// Возвращает:
//   - обновленный список метрик.
//   - ошибку types.ErrMetricInternal в случае внутренней ошибки.
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

// aggregateMetrics агрегирует входящие метрики по их ID, суммируя счетчики и обновляя показатели.
//
// Возвращает агрегированный список метрик или ошибку при неизвестном типе метрики.
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

// metricUpdateStrategies — карта стратегий обновления метрик по типу.
var metricUpdateStrategies = map[types.MetricType]func(
	oldValue types.Metrics, newValue types.Metrics,
) types.Metrics{
	types.CounterMetricType: metricUpdateCounter,
	types.GaugeMetricType:   metricUpdateGauge,
}

// metricUpdateCounter обновляет счетчик, суммируя старое и новое значения Delta.
func metricUpdateCounter(
	oldValue types.Metrics, newValue types.Metrics,
) types.Metrics {
	*newValue.Delta += *oldValue.Delta
	return newValue
}

// metricUpdateGauge обновляет показатель Gauge, заменяя старое значение на новое.
func metricUpdateGauge(
	oldValue types.Metrics, newValue types.Metrics,
) types.Metrics {
	return newValue
}
