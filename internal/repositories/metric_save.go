package repositories

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricSaveRepository interface {
	Save(ctx context.Context, metric types.Metrics) error
}

type MetricSaveContextRepository struct {
	strategy MetricSaveRepository
}

func NewMetricSaveContextRepository() *MetricSaveContextRepository {
	return &MetricSaveContextRepository{}
}

func (c *MetricSaveContextRepository) SetContext(strategy MetricSaveRepository) {
	c.strategy = strategy
}

func (c *MetricSaveContextRepository) Save(ctx context.Context, metric types.Metrics) error {
	return c.strategy.Save(ctx, metric)
}
