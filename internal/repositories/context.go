package repositories

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricSaver defines the interface for saving a metric.
type MetricSaver interface {
	Save(ctx context.Context, metric types.Metrics) error
}

// MetricContextSaveRepository uses a strategy pattern to save metrics.
type MetricContextSaveRepository struct {
	strategy MetricSaver
}

// NewMetricContextSaveRepository creates a new MetricContextSaveRepository.
func NewMetricContextSaveRepository() *MetricContextSaveRepository {
	return &MetricContextSaveRepository{}
}

// SetContext sets the saving strategy for the repository.
func (c *MetricContextSaveRepository) SetContext(strategy MetricSaver) {
	c.strategy = strategy
}

// Save saves a metric using the current strategy.
func (c *MetricContextSaveRepository) Save(ctx context.Context, metric types.Metrics) error {
	return c.strategy.Save(ctx, metric)
}

// MetricGetter defines the interface for retrieving a metric.
type MetricGetter interface {
	Get(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

// MetricContextGetRepository uses a strategy pattern to get metrics.
type MetricContextGetRepository struct {
	strategy MetricGetter
}

// NewMetricContextGetRepository creates a new MetricContextGetRepository.
func NewMetricContextGetRepository() *MetricContextGetRepository {
	return &MetricContextGetRepository{}
}

// SetContext sets the retrieval strategy for the repository.
func (c *MetricContextGetRepository) SetContext(strategy MetricGetter) {
	c.strategy = strategy
}

// Get retrieves a metric using the current strategy.
func (c *MetricContextGetRepository) Get(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
	return c.strategy.Get(ctx, id)
}

// MetricLister defines the interface for listing metrics.
type MetricLister interface {
	List(ctx context.Context) ([]*types.Metrics, error)
}

// MetricContextListRepository uses a strategy pattern to list metrics.
type MetricContextListRepository struct {
	strategy MetricLister
}

// NewMetricContextListRepository creates a new MetricContextListRepository.
func NewMetricContextListRepository() *MetricContextListRepository {
	return &MetricContextListRepository{}
}

// SetContext sets the listing strategy for the repository.
func (c *MetricContextListRepository) SetContext(strategy MetricLister) {
	c.strategy = strategy
}

// List lists metrics using the current strategy.
func (c *MetricContextListRepository) List(ctx context.Context) ([]*types.Metrics, error) {
	return c.strategy.List(ctx)
}
