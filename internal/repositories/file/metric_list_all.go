package file

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sort"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricListAllFileRepository struct {
	pathToFile string
}

func NewMetricListAllFileRepository(pathToFile string) *MetricListAllFileRepository {
	return &MetricListAllFileRepository{
		pathToFile: pathToFile,
	}
}

func (r *MetricListAllFileRepository) ListAll(
	ctx context.Context,
) ([]types.Metrics, error) {
	file, err := os.Open(r.pathToFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	metricsMap := make(map[types.MetricID]types.Metrics)
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
			ID:   metric.ID,
			Type: metric.Type,
		}
		metricsMap[key] = metric
	}

	ids := make([]types.MetricID, 0, len(metricsMap))
	for id := range metricsMap {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool {
		return ids[i].ID < ids[j].ID
	})

	metricsSlice := make([]types.Metrics, 0, len(ids))
	for _, id := range ids {
		metricsSlice = append(metricsSlice, metricsMap[id])
	}

	return metricsSlice, nil
}
