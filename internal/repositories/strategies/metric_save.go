package strategies

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricSaveRepository описывает интерфейс для сохранения метрик.
type MetricSaveRepository interface {
	// Save сохраняет метрику.
	//
	// Параметры:
	//   - ctx: контекст выполнения.
	//   - metric: структура метрики для сохранения.
	//
	// Возвращает:
	//   - ошибку, если операция не удалась.
	Save(ctx context.Context, metric types.Metrics) error
}

// MetricSaveContextRepository реализует паттерн стратегий для сохранения метрик.
//
// Позволяет динамически менять стратегию сохранения метрик.
type MetricSaveContextRepository struct {
	strategy MetricSaveRepository
}

// NewMetricSaveContextRepository создает новый экземпляр MetricSaveContextRepository.
func NewMetricSaveContextRepository() *MetricSaveContextRepository {
	return &MetricSaveContextRepository{}
}

// SetContext задает текущую стратегию сохранения метрик.
//
// Параметры:
//   - strategy: реализация интерфейса MetricSaveRepository.
func (c *MetricSaveContextRepository) SetContext(strategy MetricSaveRepository) {
	c.strategy = strategy
}

// Save сохраняет метрику, делегируя вызов текущей стратегии.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - metric: структура метрики для сохранения.
//
// Возвращает:
//   - ошибку, если операция не удалась.
func (c *MetricSaveContextRepository) Save(ctx context.Context, metric types.Metrics) error {
	return c.strategy.Save(ctx, metric)
}
