package repositories

import (
	"context"
	"sort"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

var (
	muMetricMemory *sync.RWMutex
	data           map[types.MetricID]types.Metrics
)

func init() {
	muMetricMemory = new(sync.RWMutex)
	data = make(map[types.MetricID]types.Metrics)
}

type MetricSaveMemoryRepository struct{}

func NewMetricSaveMemoryRepository() *MetricSaveMemoryRepository {
	return &MetricSaveMemoryRepository{}
}

func (r *MetricSaveMemoryRepository) Save(
	ctx context.Context,
	metric types.Metrics,
) error {
	muMetricMemory.Lock()
	defer muMetricMemory.Unlock()

	key := types.MetricID{
		ID:    metric.ID,
		MType: metric.MType,
	}

	data[key] = metric
	return nil
}

type MetricGetByIDMemoryRepository struct{}

func NewMetricGetByIDMemoryRepository() *MetricGetByIDMemoryRepository {
	return &MetricGetByIDMemoryRepository{}
}

func (r *MetricGetByIDMemoryRepository) Get(
	ctx context.Context, id types.MetricID,
) (*types.Metrics, error) {
	muMetricMemory.RLock()
	defer muMetricMemory.RUnlock()

	metric, exists := data[id]
	if !exists {
		return nil, nil
	}

	return &metric, nil
}

type MetricListAllMemoryRepository struct{}

func NewMetricListAllMemoryRepository() *MetricListAllMemoryRepository {
	return &MetricListAllMemoryRepository{}
}

func (r *MetricListAllMemoryRepository) List(
	ctx context.Context,
) ([]*types.Metrics, error) {
	muMetricMemory.RLock()
	defer muMetricMemory.RUnlock()

	metrics := make([]*types.Metrics, 0, len(data))
	for _, m := range data {
		copy := m
		metrics = append(metrics, &copy)
	}

	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	return metrics, nil
}
