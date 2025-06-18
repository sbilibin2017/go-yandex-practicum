package services

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricGetByIDRepository описывает интерфейс репозитория для получения метрики по ID.
type MetricGetByIDRepository interface {
	// GetByID возвращает метрику по её уникальному идентификатору.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - id: идентификатор метрики.
	//
	// Возвращает:
	//   - указатель на найденную метрику, если она существует.
	//   - ошибку, если произошла внутренняя ошибка.
	GetByID(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

// MetricGetService реализует логику получения метрик по ID.
type MetricGetService struct {
	mfor MetricGetByIDRepository
}

// NewMetricGetService создает новый сервис для получения метрик.
//
// Параметры:
//   - mfor: реализация интерфейса MetricGetByIDRepository.
//
// Возвращает:
//   - новый экземпляр MetricGetService.
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
		return nil, err
	}
	return currentMetric, nil
}
