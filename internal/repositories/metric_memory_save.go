package repositories

import (
	"context"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricMemorySaveRepository struct {
	data map[types.MetricID]types.Metrics
	mu   sync.RWMutex
}

func NewMetricMemorySaveRepository(
	data map[types.MetricID]types.Metrics,
) *MetricMemorySaveRepository {
	return &MetricMemorySaveRepository{
		data: data,
	}
}

func (r *MetricMemorySaveRepository) Save(
	ctx context.Context, metrics types.Metrics,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[metrics.MetricID] = metrics
	return nil
}
