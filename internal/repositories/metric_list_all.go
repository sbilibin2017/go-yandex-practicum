package repositories

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricListAllRepository interface {
	ListAll(ctx context.Context) ([]types.Metrics, error)
}

type MetricListAllContextRepository struct {
	strategy MetricListAllRepository
}

func NewMetricListAllContextRepository() *MetricListAllContextRepository {
	return &MetricListAllContextRepository{}
}

func (c *MetricListAllContextRepository) SetContext(strategy MetricListAllRepository) {
	c.strategy = strategy
}

func (c *MetricListAllContextRepository) ListAll(ctx context.Context) ([]types.Metrics, error) {
	return c.strategy.ListAll(ctx)
}
