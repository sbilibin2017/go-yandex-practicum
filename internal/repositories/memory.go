package repositories

import (
	"context"
	"sort"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

var (
	// muMemory protects concurrent access to the in-memory data map.
	muMemory *sync.RWMutex

	// data stores metrics in memory indexed by MetricID.
	data map[types.MetricID]types.Metrics
)

func init() {
	muMemory = new(sync.RWMutex)
	data = make(map[types.MetricID]types.Metrics)
}

// MetricMemorySaveRepository provides methods to save metrics in memory.
type MetricMemorySaveRepository struct{}

// NewMetricMemorySaveRepository creates a new MetricMemorySaveRepository.
func NewMetricMemorySaveRepository() *MetricMemorySaveRepository {
	return &MetricMemorySaveRepository{}
}

// Save stores the given metric in the in-memory map.
// It acquires a write lock to ensure concurrent safety.
func (r *MetricMemorySaveRepository) Save(ctx context.Context, metric types.Metrics) error {
	muMemory.Lock()
	defer muMemory.Unlock()

	key := types.MetricID{
		ID:   metric.ID,
		Type: metric.Type,
	}

	data[key] = metric
	return nil
}

// MetricMemoryGetRepository provides methods to get metrics from memory.
type MetricMemoryGetRepository struct{}

// NewMetricMemoryGetRepository creates a new MetricMemoryGetRepository.
func NewMetricMemoryGetRepository() *MetricMemoryGetRepository {
	return &MetricMemoryGetRepository{}
}

// Get retrieves a metric by its MetricID from the in-memory map.
// Returns nil if no metric with the given ID exists.
func (r *MetricMemoryGetRepository) Get(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
	muMemory.RLock()
	defer muMemory.RUnlock()

	metric, exists := data[id]
	if !exists {
		return nil, nil
	}

	return &metric, nil
}

// MetricMemoryListRepository provides methods to list all metrics from memory.
type MetricMemoryListRepository struct{}

// NewMetricMemoryListRepository creates a new MetricMemoryListRepository.
func NewMetricMemoryListRepository() *MetricMemoryListRepository {
	return &MetricMemoryListRepository{}
}

// List returns all stored metrics as a slice sorted by metric ID.
// It acquires a read lock during the operation.
func (r *MetricMemoryListRepository) List(ctx context.Context) ([]*types.Metrics, error) {
	muMemory.RLock()
	defer muMemory.RUnlock()

	var metrics []*types.Metrics
	for _, m := range data {
		c := m
		metrics = append(metrics, &c)
	}

	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	return metrics, nil
}
