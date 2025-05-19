package services

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricListAllRepository описывает интерфейс репозитория для получения всех метрик.
type MetricListAllRepository interface {
	// ListAll возвращает список всех метрик.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//
	// Возвращает:
	//   - срез метрик.
	//   - ошибку в случае неудачи.
	ListAll(ctx context.Context) ([]types.Metrics, error)
}

// MetricListAllService реализует бизнес-логику получения всех метрик.
type MetricListAllService struct {
	mlar MetricListAllRepository
}

// NewMetricListAllService создает новый сервис для получения всех метрик.
//
// Параметры:
//   - mlar: реализация интерфейса MetricListAllRepository.
//
// Возвращает:
//   - новый экземпляр MetricListAllService.
func NewMetricListAllService(
	mlar MetricListAllRepository,
) *MetricListAllService {
	return &MetricListAllService{
		mlar: mlar,
	}
}

// ListAll возвращает список всех метрик.
//
// Параметры:
//   - ctx: контекст выполнения.
//
// Возвращает:
//   - срез метрик, если они есть.
//   - nil, если метрик нет.
//   - ошибку types.ErrMetricInternal в случае внутренней ошибки.
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
