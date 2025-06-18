package strategies

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricListAllRepository определяет интерфейс для получения всех метрик.
type MetricListAllRepository interface {
	// ListAll возвращает все метрики.
	//
	// Параметры:
	//  - ctx: контекст выполнения.
	//
	// Возвращает:
	//  - срез метрик ([]types.Metrics).
	//  - ошибку, если она возникла.
	ListAll(ctx context.Context) ([]types.Metrics, error)
}

// MetricListAllContextRepository реализует паттерн стратегия
// для вызова метода ListAll через выбранную стратегию.
type MetricListAllContextRepository struct {
	strategy MetricListAllRepository
}

// NewMetricListAllContextRepository создает новый экземпляр MetricListAllContextRepository.
func NewMetricListAllContextRepository() *MetricListAllContextRepository {
	return &MetricListAllContextRepository{}
}

// SetContext задает стратегию для получения метрик.
//
// Параметры:
//   - strategy: реализация интерфейса MetricListAllRepository.
func (c *MetricListAllContextRepository) SetContext(strategy MetricListAllRepository) {
	c.strategy = strategy
}

// ListAll вызывает метод ListAll выбранной стратегии.
//
// Параметры:
//   - ctx: контекст выполнения.
//
// Возвращает:
//   - срез метрик ([]types.Metrics).
//   - ошибку, если она возникла.
func (c *MetricListAllContextRepository) ListAll(ctx context.Context) ([]types.Metrics, error) {
	return c.strategy.ListAll(ctx)
}
