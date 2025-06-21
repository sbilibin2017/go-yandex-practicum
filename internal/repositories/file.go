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

var muFile = new(sync.RWMutex)

//
// MetricFileSaveRepository
//

type MetricFileSaveRepository struct {
	metricFilePath string
}

type MetricFileSaveRepositoryOption func(*MetricFileSaveRepository)

func WithMetricFileSaveRepositoryPath(path string) MetricFileSaveRepositoryOption {
	return func(r *MetricFileSaveRepository) {
		r.metricFilePath = path
	}
}

func NewMetricFileSaveRepository(opts ...MetricFileSaveRepositoryOption) *MetricFileSaveRepository {
	repo := &MetricFileSaveRepository{}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

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

//
// MetricFileGetRepository
//

type MetricFileGetRepository struct {
	metricFilePath string
}

type MetricFileGetRepositoryOption func(*MetricFileGetRepository)

func WithMetricFileGetRepositoryPath(path string) MetricFileGetRepositoryOption {
	return func(r *MetricFileGetRepository) {
		r.metricFilePath = path
	}
}

func NewMetricFileGetRepository(opts ...MetricFileGetRepositoryOption) *MetricFileGetRepository {
	repo := &MetricFileGetRepository{}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

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

		if metric.ID == id.ID && metric.MType == id.MType {
			return &metric, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}

//
// MetricFileListRepository
//

type MetricFileListRepository struct {
	metricFilePath string
}

type MetricFileListRepositoryOption func(*MetricFileListRepository)

func WithMetricFileListRepositoryPath(path string) MetricFileListRepositoryOption {
	return func(r *MetricFileListRepository) {
		r.metricFilePath = path
	}
}

func NewMetricFileListRepository(opts ...MetricFileListRepositoryOption) *MetricFileListRepository {
	repo := &MetricFileListRepository{}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

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
		key := types.MetricID{ID: m.ID, MType: m.MType}
		mCopy := m
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
