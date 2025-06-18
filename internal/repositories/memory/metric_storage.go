package memory

import (
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

var (
	data map[types.MetricID]types.Metrics
	mu   *sync.RWMutex
)

func init() {
	data = make(map[types.MetricID]types.Metrics)
	mu = &sync.RWMutex{}
}
