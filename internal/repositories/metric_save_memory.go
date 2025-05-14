package repositories

import (
	"context"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricSaveMemoryRepository struct {
	data map[types.MetricID]types.Metrics
	mu   sync.RWMutex
}

func NewMetricSaveMemoryRepository(
	data map[types.MetricID]types.Metrics,
) *MetricSaveMemoryRepository {
	return &MetricSaveMemoryRepository{
		data: data,
	}
}

func (r *MetricSaveMemoryRepository) Save(
	ctx context.Context, metric types.Metrics,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[metric.MetricID] = metric
	return nil
}
