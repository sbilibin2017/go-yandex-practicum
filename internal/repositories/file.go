package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

var (
	muMetricFile *sync.RWMutex
)

func init() {
	muMetricFile = new(sync.RWMutex)
}

type MetricSaveFileRepository struct {
	pathToFile string
}

func NewMetricSaveFileRepository(pathToFile string) *MetricSaveFileRepository {
	return &MetricSaveFileRepository{
		pathToFile: pathToFile,
	}
}

func (r *MetricSaveFileRepository) Save(ctx context.Context, metric types.Metrics) error {
	muMetricFile.Lock()
	defer muMetricFile.Unlock()

	origFile, err := os.Open(r.pathToFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if origFile != nil {
		defer origFile.Close()
	}

	tmpFilePath := r.pathToFile + ".tmp"
	tmpFile, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	decoder := json.NewDecoder(origFile)
	encoder := json.NewEncoder(tmpFile)

	found := false

	for {
		var m types.Metrics
		err := decoder.Decode(&m)
		if err != nil {
			break
		}

		if m.ID == metric.ID && m.MType == metric.MType {
			if err := encoder.Encode(metric); err != nil {
				return err
			}
			found = true
		} else {
			if err := encoder.Encode(m); err != nil {
				return err
			}
		}
	}

	if !found {
		if err := encoder.Encode(metric); err != nil {
			return err
		}
	}

	if err := os.Rename(tmpFilePath, r.pathToFile); err != nil {
		return err
	}

	return nil
}

type MetricListAllFileRepository struct {
	pathToFile string
}

func NewMetricListAllFileRepository(pathToFile string) *MetricListAllFileRepository {
	return &MetricListAllFileRepository{
		pathToFile: pathToFile,
	}
}

func (r *MetricListAllFileRepository) List(
	ctx context.Context,
) ([]*types.Metrics, error) {
	muMetricFile.Lock()
	defer muMetricFile.Unlock()

	file, err := os.Open(r.pathToFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	metricsMap := make(map[types.MetricID]*types.Metrics)
	decoder := json.NewDecoder(file)

	for {
		var metric types.Metrics
		err := decoder.Decode(&metric)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		key := types.MetricID{
			ID:    metric.ID,
			MType: metric.MType,
		}
		metricsMap[key] = &metric
	}

	ids := make([]types.MetricID, 0, len(metricsMap))
	for id := range metricsMap {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool {
		return ids[i].ID < ids[j].ID
	})

	metricsSlice := make([]*types.Metrics, 0, len(ids))
	for _, id := range ids {
		metricsSlice = append(metricsSlice, metricsMap[id])
	}

	return metricsSlice, nil
}

type MetricGetByIDFileRepository struct {
	pathToFile string
}

func NewMetricGetByIDFileRepository(pathToFile string) *MetricGetByIDFileRepository {
	return &MetricGetByIDFileRepository{
		pathToFile: pathToFile,
	}
}

func (r *MetricGetByIDFileRepository) Get(
	ctx context.Context, id types.MetricID,
) (*types.Metrics, error) {
	muMetricFile.Lock()
	defer muMetricFile.Unlock()

	file, err := os.Open(r.pathToFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	for {
		var metric types.Metrics
		if err := decoder.Decode(&metric); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		if metric.ID == id.ID && metric.MType == id.MType {
			return &metric, nil
		}
	}

	return nil, errors.New("metric not found")
}
