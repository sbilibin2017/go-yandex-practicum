package repositories

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricGetByIDRepository interface {
	GetByID(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

type MetricGetByIDContextRepository struct {
	strategy MetricGetByIDRepository
}

func NewMetricGetByIDContextRepository() *MetricGetByIDContextRepository {
	return &MetricGetByIDContextRepository{}
}

func (c *MetricGetByIDContextRepository) SetContext(strategy MetricGetByIDRepository) {
	c.strategy = strategy
}

func (c *MetricGetByIDContextRepository) GetByID(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
	return c.strategy.GetByID(ctx, id)
}
