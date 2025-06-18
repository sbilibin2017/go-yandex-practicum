package file

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricGetByIDFileRepository struct {
	pathToFile string
}

func NewMetricGetByIDFileRepository(pathToFile string) *MetricGetByIDFileRepository {
	return &MetricGetByIDFileRepository{
		pathToFile: pathToFile,
	}
}

func (r *MetricGetByIDFileRepository) GetByID(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
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

		if metric.ID == id.ID && metric.Type == id.Type {
			return &metric, nil
		}
	}

	return nil, errors.New("metric not found")
}
