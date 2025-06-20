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
	ctx context.Context, metrics []*types.Metrics,
) ([]*types.Metrics, error) {
	updatedMetrics := make(map[types.MetricID]*types.Metrics)

	for _, metric := range metrics {
		switch metric.MType {
		case types.Counter:
			currentMetric, err := svc.mfor.GetByID(ctx, types.MetricID{ID: metric.ID, MType: metric.MType})
			if err != nil {
				return nil, err
			}

			var current int64
			if currentMetric != nil && currentMetric.Delta != nil {
				current = *currentMetric.Delta
			}

			if metric.Delta == nil {
				var initialDelta int64 = 0
				metric.Delta = &initialDelta
			}

			*metric.Delta += current
		}

		if err := svc.msr.Save(ctx, *metric); err != nil {
			return nil, err
		}

		id := types.MetricID{ID: metric.ID, MType: metric.MType}
		updatedMetrics[id] = metric
	}

	result := make([]*types.Metrics, 0, len(updatedMetrics))
	for _, metric := range updatedMetrics {
		result = append(result, metric)
	}

	return result, nil
}
