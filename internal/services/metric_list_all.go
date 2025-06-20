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
	ListAll(ctx context.Context) ([]*types.Metrics, error)
}

// MetricListAllService реализует бизнес-логику получения всех метрик.
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
) ([]*types.Metrics, error) {
	metrics, err := svc.mlar.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	if len(metrics) == 0 {
		return nil, nil
	}
	return metrics, nil
}
