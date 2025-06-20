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

type MetricListAllRepository interface {
	List(ctx context.Context) ([]*types.Metrics, error)
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

func (c *MetricListAllContextRepository) List(ctx context.Context) ([]*types.Metrics, error) {
	return c.strategy.List(ctx)
}

type MetricGetByIDRepository interface {
	Get(ctx context.Context, id types.MetricID) (*types.Metrics, error)
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

func (c *MetricGetByIDContextRepository) Get(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
	return c.strategy.Get(ctx, id)
}
