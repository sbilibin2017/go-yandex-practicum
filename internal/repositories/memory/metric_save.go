package memory

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricSaveMemoryRepository struct{}

func NewMetricSaveMemoryRepository() *MetricSaveMemoryRepository {
	return &MetricSaveMemoryRepository{}
}

func (r *MetricSaveMemoryRepository) Save(
	ctx context.Context, metric types.Metrics,
) error {
	mu.Lock()
	defer mu.Unlock()

	key := types.MetricID{
		ID:   metric.ID,
		Type: metric.Type,
	}

	data[key] = metric
	return nil
}
