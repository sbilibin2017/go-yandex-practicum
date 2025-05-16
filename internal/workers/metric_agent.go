package workers

import (
	"context"
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

type Result struct {
	data types.Metrics
	err  error
}

type MetricFacade interface {
	Updates(ctx context.Context, metrics []types.Metrics) error
}

type Semaphore interface {
	Acquire(ctx context.Context, n int64) error
	Release(n int64)
}

type MetricAgentWorker struct {
	facade         MetricFacade
	sem            Semaphore
	pollInterval   int
	workerCount    int
	batchSize      int
	reportInterval int
}

func NewMetricAgentWorker(
	facade MetricFacade,
	sem Semaphore,
	pollInterval int,
	workerCount int,
	batchSize int,
	reportInterval int,
) *MetricAgentWorker {
	return &MetricAgentWorker{
		facade:         facade,
		sem:            sem,
		pollInterval:   pollInterval,
		workerCount:    workerCount,
		batchSize:      batchSize,
		reportInterval: reportInterval,
	}
}

func (m *MetricAgentWorker) Start(ctx context.Context) error {
	return startMetricAgent(
		ctx,
		m.facade,
		m.sem,
		m.pollInterval,
		m.workerCount,
		m.batchSize,
		m.reportInterval,
	)
}

func startMetricAgent(
	ctx context.Context,
	facade MetricFacade,
	sem Semaphore,
	pollInterval int,
	workerCount int,
	batchSize int,
	reportInterval int,
) error {
	runtimeMetricsCh := generatorRuntimeMetrics(ctx, pollInterval)
	gopsutilMetricsCh := generatorGopsutilMetrics(ctx, pollInterval)
	mergedMetricsCh := fanIn(ctx, runtimeMetricsCh, gopsutilMetricsCh)

	resultCh := make(chan Result, 10*batchSize)
	go fanOutWorkerPool(
		ctx,
		facade,
		sem,
		mergedMetricsCh,
		workerCount,
		batchSize,
		reportInterval,
		resultCh,
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-resultCh:
			if !ok {
				return nil
			}
			if res.err != nil {
				logger.Log.Error("Error updating metric",
					zap.String("metricID", res.data.MetricID.ID),
					zap.Error(res.err),
				)
			}
		}
	}
}

func generatorRuntimeMetrics(ctx context.Context, pollInterval int) chan types.Metrics {
	metricsChan := make(chan types.Metrics, 100)

	go func() {
		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()
		defer close(metricsChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, m := range generateRuntimeGaugeMetrics(ctx) {
					select {
					case metricsChan <- m:
					case <-ctx.Done():
						return
					}
				}
				for _, m := range generateRuntimeCounterMetrics(ctx) {
					select {
					case metricsChan <- m:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return metricsChan
}

func generatorGopsutilMetrics(ctx context.Context, pollInterval int) chan types.Metrics {
	metricsChan := make(chan types.Metrics, 100)

	go func() {
		defer close(metricsChan)

		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, m := range generateGopsutilGaugeMetrics(ctx) {
					select {
					case metricsChan <- m:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return metricsChan
}
func generateRuntimeGaugeMetrics(_ context.Context) []types.Metrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var metrics []types.Metrics

	v1 := float64(memStats.Alloc)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "Alloc", Type: types.GaugeMetricType},
		Value:    &v1,
	})

	v2 := float64(memStats.BuckHashSys)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "BuckHashSys", Type: types.GaugeMetricType},
		Value:    &v2,
	})

	v3 := float64(memStats.Frees)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "Frees", Type: types.GaugeMetricType},
		Value:    &v3,
	})

	v4 := memStats.GCCPUFraction
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "GCCPUFraction", Type: types.GaugeMetricType},
		Value:    &v4,
	})

	v5 := float64(memStats.GCSys)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "GCSys", Type: types.GaugeMetricType},
		Value:    &v5,
	})

	v6 := float64(memStats.HeapAlloc)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "HeapAlloc", Type: types.GaugeMetricType},
		Value:    &v6,
	})

	v7 := float64(memStats.HeapIdle)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "HeapIdle", Type: types.GaugeMetricType},
		Value:    &v7,
	})

	v8 := float64(memStats.HeapInuse)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "HeapInuse", Type: types.GaugeMetricType},
		Value:    &v8,
	})

	v9 := float64(memStats.HeapObjects)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "HeapObjects", Type: types.GaugeMetricType},
		Value:    &v9,
	})

	v10 := float64(memStats.HeapReleased)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "HeapReleased", Type: types.GaugeMetricType},
		Value:    &v10,
	})

	v11 := float64(memStats.HeapSys)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "HeapSys", Type: types.GaugeMetricType},
		Value:    &v11,
	})

	v12 := float64(memStats.LastGC)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "LastGC", Type: types.GaugeMetricType},
		Value:    &v12,
	})

	v13 := float64(memStats.Lookups)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "Lookups", Type: types.GaugeMetricType},
		Value:    &v13,
	})

	v14 := float64(memStats.MCacheInuse)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "MCacheInuse", Type: types.GaugeMetricType},
		Value:    &v14,
	})

	v15 := float64(memStats.MCacheSys)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "MCacheSys", Type: types.GaugeMetricType},
		Value:    &v15,
	})

	v16 := float64(memStats.MSpanInuse)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "MSpanInuse", Type: types.GaugeMetricType},
		Value:    &v16,
	})

	v17 := float64(memStats.MSpanSys)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "MSpanSys", Type: types.GaugeMetricType},
		Value:    &v17,
	})

	v18 := float64(memStats.Mallocs)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "Mallocs", Type: types.GaugeMetricType},
		Value:    &v18,
	})

	v19 := float64(memStats.NextGC)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "NextGC", Type: types.GaugeMetricType},
		Value:    &v19,
	})

	v20 := float64(memStats.NumForcedGC)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "NumForcedGC", Type: types.GaugeMetricType},
		Value:    &v20,
	})

	v21 := float64(memStats.NumGC)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "NumGC", Type: types.GaugeMetricType},
		Value:    &v21,
	})

	v22 := float64(memStats.OtherSys)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "OtherSys", Type: types.GaugeMetricType},
		Value:    &v22,
	})

	v23 := float64(memStats.PauseTotalNs)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "PauseTotalNs", Type: types.GaugeMetricType},
		Value:    &v23,
	})

	v24 := float64(memStats.StackInuse)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "StackInuse", Type: types.GaugeMetricType},
		Value:    &v24,
	})

	v25 := float64(memStats.StackSys)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "StackSys", Type: types.GaugeMetricType},
		Value:    &v25,
	})

	v26 := float64(memStats.Sys)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "Sys", Type: types.GaugeMetricType},
		Value:    &v26,
	})

	v27 := float64(memStats.TotalAlloc)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "TotalAlloc", Type: types.GaugeMetricType},
		Value:    &v27,
	})

	v28 := rand.Float64()
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "RandomValue", Type: types.GaugeMetricType},
		Value:    &v28,
	})

	return metrics
}

func generateRuntimeCounterMetrics(_ context.Context) []types.Metrics {
	var metrics []types.Metrics

	v1 := int64(1)
	metrics = append(metrics, types.Metrics{
		MetricID: types.MetricID{ID: "PollCount", Type: types.CounterMetricType},
		Delta:    &v1,
	})

	return metrics
}

func generateGopsutilGaugeMetrics(ctx context.Context) []types.Metrics {
	var metrics []types.Metrics

	if vmStat, err := mem.VirtualMemory(); err == nil {
		total := float64(vmStat.Total)
		free := float64(vmStat.Free)

		metrics = append(metrics, types.Metrics{
			MetricID: types.MetricID{ID: "TotalMemory", Type: types.GaugeMetricType},
			Value:    &total,
		})
		metrics = append(metrics, types.Metrics{
			MetricID: types.MetricID{ID: "FreeMemory", Type: types.GaugeMetricType},
			Value:    &free,
		})
	}

	if cpuPercents, err := cpu.PercentWithContext(ctx, 0, true); err == nil {
		for i, percent := range cpuPercents {
			p := percent
			id := "CPUutilization" + strconv.Itoa(i)

			metrics = append(metrics, types.Metrics{
				MetricID: types.MetricID{ID: id, Type: types.GaugeMetricType},
				Value:    &p,
			})
		}
	}

	return metrics
}

func fanIn(ctx context.Context, chans ...<-chan types.Metrics) <-chan types.Metrics {
	out := make(chan types.Metrics)

	var wg sync.WaitGroup
	wg.Add(len(chans))

	for _, ch := range chans {
		go func(c <-chan types.Metrics) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case m, ok := <-c:
					if !ok {
						return
					}
					select {
					case out <- m:
					case <-ctx.Done():
						return
					}
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func fanOutWorker(
	ctx context.Context,
	facade MetricFacade,
	sem Semaphore,
	inputCh <-chan types.Metrics,
	batchSize int,
	reportInterval int,
	resultCh chan<- Result,
) {
	batch := make([]types.Metrics, 0, batchSize)
	dur := time.Duration(reportInterval) * time.Second
	timer := time.NewTimer(dur)
	defer timer.Stop()

	sendBatch := func() {
		if len(batch) == 0 {
			return
		}

		if err := sem.Acquire(ctx, 1); err != nil {
			return
		}

		batchCopy := append([]types.Metrics(nil), batch...)
		batch = batch[:0]

		go func() {
			defer sem.Release(1)

			err := facade.Updates(ctx, batchCopy)
			if err != nil {
				for _, m := range batchCopy {
					resultCh <- Result{
						data: m,
						err:  err,
					}
				}
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			sendBatch()
			return

		case m, ok := <-inputCh:
			if !ok {
				sendBatch()
				return
			}
			batch = append(batch, m)

			if len(batch) >= batchSize {
				sendBatch()
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(dur)
			}

		case <-timer.C:
			sendBatch()
			timer.Reset(dur)
		}
	}
}

func fanOutWorkerPool(
	ctx context.Context,
	facade MetricFacade,
	sem Semaphore,
	inputCh <-chan types.Metrics,
	workerCount int,
	batchSize int,
	reportInterval int,
	resultCh chan<- Result,
) {
	var wg sync.WaitGroup
	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			fanOutWorker(ctx, facade, sem, inputCh, batchSize, reportInterval, resultCh)
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()
}
