package repositories

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sort"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

var (
	// muFile protects concurrent access to file operations.
	muFile *sync.RWMutex
)

func init() {
	muFile = new(sync.RWMutex)
}

// FileConfigOption defines a functional option for configuring file repositories.
type FileConfigOption func(*FileConfig)

// FileConfig holds configuration options for file repositories.
type FileConfig struct {
	MetricFilePath string
}

// NewFileConfig applies functional options and returns a new FileConfig.
func NewFileConfig(opts ...FileConfigOption) *FileConfig {
	cfg := &FileConfig{
		MetricFilePath: "",
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// WithFilePath sets the file path for the repository.
func WithFilePath(path string) FileConfigOption {
	return func(cfg *FileConfig) {
		cfg.MetricFilePath = path
	}
}

// MetricFileSaveRepository saves metrics into a JSON file, overwriting it.
type MetricFileSaveRepository struct {
	metricFilePath string
}

// NewMetricFileSaveRepository creates a new MetricFileSaveRepository using functional options.
func NewMetricFileSaveRepository(opts ...FileConfigOption) *MetricFileSaveRepository {
	cfg := NewFileConfig(opts...)
	return &MetricFileSaveRepository{metricFilePath: cfg.MetricFilePath}
}

// Save writes the given metric to the file, overwriting existing content.
// It acquires a write lock to ensure safe concurrent access.
func (r *MetricFileSaveRepository) Save(ctx context.Context, metric types.Metrics) error {
	muFile.Lock()
	defer muFile.Unlock()

	file, err := os.OpenFile(r.metricFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	return enc.Encode(metric)
}

// MetricFileGetRepository reads metrics from a JSON file.
type MetricFileGetRepository struct {
	metricFilePath string
}

// NewMetricFileGetRepository creates a new MetricFileGetRepository using functional options.
func NewMetricFileGetRepository(opts ...FileConfigOption) *MetricFileGetRepository {
	cfg := NewFileConfig(opts...)
	return &MetricFileGetRepository{metricFilePath: cfg.MetricFilePath}
}

// Get reads the file and attempts to find a metric matching the provided id.
// It returns (nil, nil) if the file does not exist or the metric is not found.
// Uses a read lock for safe concurrent access.
func (r *MetricFileGetRepository) Get(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
	muFile.RLock()
	defer muFile.RUnlock()

	file, err := os.Open(r.metricFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var metric types.Metrics
		if err := json.Unmarshal(scanner.Bytes(), &metric); err != nil {
			return nil, err
		}
		key := types.MetricID{ID: metric.ID, MType: metric.MType}
		if key == id {
			return &metric, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}

// MetricFileListRepository lists all metrics from a JSON file.
type MetricFileListRepository struct {
	metricFilePath string
}

// NewMetricFileListRepository creates a new MetricFileListRepository using functional options.
func NewMetricFileListRepository(opts ...FileConfigOption) *MetricFileListRepository {
	cfg := NewFileConfig(opts...)
	return &MetricFileListRepository{metricFilePath: cfg.MetricFilePath}
}

// List reads all metrics from the file and returns them sorted by ID.
// It returns (nil, nil) if the file does not exist.
// Uses a read lock for safe concurrent access.
func (r *MetricFileListRepository) List(ctx context.Context) ([]*types.Metrics, error) {
	muFile.RLock()
	defer muFile.RUnlock()

	file, err := os.Open(r.metricFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	metricsMap := make(map[types.MetricID]*types.Metrics)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var m types.Metrics
		if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {
			return nil, err
		}

		mCopy := m
		key := types.MetricID{ID: mCopy.ID, MType: mCopy.MType}
		metricsMap[key] = &mCopy
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	metricsSlice := make([]*types.Metrics, 0, len(metricsMap))
	for _, m := range metricsMap {
		metricsSlice = append(metricsSlice, m)
	}

	sort.SliceStable(metricsSlice, func(i, j int) bool {
		return metricsSlice[i].ID < metricsSlice[j].ID
	})

	return metricsSlice, nil
}
