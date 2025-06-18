package memory

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricGetByIDMemoryRepository struct{}

// NewMetricGetByIDMemoryRepository creates a new instance.
func NewMetricGetByIDMemoryRepository() *MetricGetByIDMemoryRepository {
	return &MetricGetByIDMemoryRepository{}
}

// GetByID returns metric by id or nil if not found.
func (r *MetricGetByIDMemoryRepository) GetByID(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
	mu.RLock()
	defer mu.RUnlock()

	metric, exists := data[id]
	if !exists {
		return nil, nil
	}

	return &metric, nil
}
