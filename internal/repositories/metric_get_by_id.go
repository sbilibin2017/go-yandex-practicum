package repositories

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricGetByIDRepository определяет интерфейс для получения метрики по её идентификатору.
type MetricGetByIDRepository interface {
	// GetByID возвращает метрику по заданному идентификатору.
	//
	// Параметры:
	//  - ctx: контекст выполнения запроса.
	//  - id: идентификатор метрики.
	//
	// Возвращает:
	//  - указатель на найденную метрику, либо nil, если метрика не найдена.
	//  - ошибку, если возникли проблемы при выполнении запроса.
	GetByID(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

// MetricGetByIDContextRepository предоставляет возможность динамически
// менять стратегию получения метрик по идентификатору.
type MetricGetByIDContextRepository struct {
	strategy MetricGetByIDRepository
}

// NewMetricGetByIDContextRepository создает новый экземпляр MetricGetByIDContextRepository.
func NewMetricGetByIDContextRepository() *MetricGetByIDContextRepository {
	return &MetricGetByIDContextRepository{}
}

// SetContext устанавливает текущую стратегию получения метрик.
//
// Параметры:
//   - strategy: объект, реализующий интерфейс MetricGetByIDRepository.
func (c *MetricGetByIDContextRepository) SetContext(strategy MetricGetByIDRepository) {
	c.strategy = strategy
}

// GetByID вызывает метод GetByID установленной стратегии.
//
// Параметры:
//   - ctx: контекст выполнения запроса.
//   - id: идентификатор метрики.
//
// Возвращает:
//   - указатель на найденную метрику, либо nil, если метрика не найдена.
//   - ошибку, если возникли проблемы при выполнении запроса.
func (c *MetricGetByIDContextRepository) GetByID(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
	return c.strategy.GetByID(ctx, id)
}
