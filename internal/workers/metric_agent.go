package workers

import (
	"context"
	"math/rand"
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

type MetricUpdater interface {
	Updates(ctx context.Context, metrics []*types.Metrics) error
}

type result struct {
	Data []*types.Metrics
	Err  error
}

func NewMetricAgentWorker(
	updater MetricUpdater,
	pollInterval int,
	reportInterval int,
	batchSize int,
	rateLimit int,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return startMetricAgent(ctx, updater, pollInterval, reportInterval, batchSize, rateLimit)
	}
}

func startMetricAgent(
	ctx context.Context,
	updater MetricUpdater,
	pollInterval int,
	reportInterval int,
	batchSize int,
	rateLimit int,
) error {

	metricsCh := startMetricsPolling(ctx, pollInterval)
	resultsCh := startMetricsReporting(ctx, updater, reportInterval, metricsCh, batchSize, rateLimit)
	return logResults(ctx, resultsCh)
}

func startMetricsPolling(ctx context.Context, pollInterval int) <-chan types.Metrics {
	out := make(chan types.Metrics, 100)
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)

	go func() {
		defer ticker.Stop()
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runtimeCh := generatorMetrics(ctx, getRuntimeGaugeMetrics())
				counterCh := generatorMetrics(ctx, getRuntimeCounterMetrics())
				gopsutilCh := generatorMetrics(ctx, getGoputilMetrics(ctx))

				for metric := range fanInMetrics(ctx, runtimeCh, counterCh, gopsutilCh) {
					out <- metric
				}
			}
		}
	}()

	return out
}

func startMetricsReporting(
	ctx context.Context,
	updater MetricUpdater,
	reportInterval int,
	in <-chan types.Metrics,
	batchSize int,
	rateLimit int,
) <-chan result {
	resultsCh := make(chan result, 100)
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)

	go func() {
		defer ticker.Stop()
		defer close(resultsCh)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				results := workerPoolMetricsUpdate(ctx, updater, in, batchSize, rateLimit)
				for res := range results {
					resultsCh <- res
				}
			}
		}
	}()

	return resultsCh
}

func workerMetricsUpdate(
	ctx context.Context,
	updater MetricUpdater,
	jobs <-chan types.Metrics,
	batchSize int,
) <-chan result {
	results := make(chan result, 100)

	go func() {
		defer close(results)

		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-jobs:
				if !ok {
					return
				}

				batch := []*types.Metrics{&m}

			collectLoop:
				for len(batch) < batchSize {
					m, ok := <-jobs
					if !ok {
						break collectLoop
					}
					batch = append(batch, &m)
				}

				err := updater.Updates(ctx, batch)

				results <- result{
					Data: batch,
					Err:  err,
				}
			}
		}
	}()

	return results
}

func workerPoolMetricsUpdate(
	ctx context.Context,
	updater MetricUpdater,
	jobs <-chan types.Metrics,
	batchSize int,
	rateLimit int,
) chan result {
	results := make(chan result)
	var wg sync.WaitGroup
	wg.Add(rateLimit)

	for i := 0; i < rateLimit; i++ {
		go func() {
			defer wg.Done()
			workerResults := workerMetricsUpdate(ctx, updater, jobs, batchSize)
			for r := range workerResults {
				results <- r
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func logResults(ctx context.Context, results <-chan result) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-results:
			if !ok {
				return ctx.Err()
			}
			if res.Err != nil {
				logger.Log.Error("worker pool task error", zap.Error(res.Err), zap.Any("data", res.Data))
			} else {
				logger.Log.Info("worker pool task success", zap.Any("data", res.Data))
			}
		}
	}
}

func getRuntimeGaugeMetrics() []types.Metrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	v1 := float64(memStats.Alloc)
	v2 := float64(memStats.BuckHashSys)
	v3 := float64(memStats.Frees)
	v4 := memStats.GCCPUFraction
	v5 := float64(memStats.GCSys)
	v6 := float64(memStats.HeapAlloc)
	v7 := float64(memStats.HeapIdle)
	v8 := float64(memStats.HeapInuse)
	v9 := float64(memStats.HeapObjects)
	v10 := float64(memStats.HeapReleased)
	v11 := float64(memStats.HeapSys)
	v12 := float64(memStats.LastGC)
	v13 := float64(memStats.Lookups)
	v14 := float64(memStats.MCacheInuse)
	v15 := float64(memStats.MCacheSys)
	v16 := float64(memStats.MSpanInuse)
	v17 := float64(memStats.MSpanSys)
	v18 := float64(memStats.Mallocs)
	v19 := float64(memStats.NextGC)
	v20 := float64(memStats.NumForcedGC)
	v21 := float64(memStats.NumGC)
	v22 := float64(memStats.OtherSys)
	v23 := float64(memStats.PauseTotalNs)
	v24 := float64(memStats.StackInuse)
	v25 := float64(memStats.StackSys)
	v26 := float64(memStats.Sys)
	v27 := float64(memStats.TotalAlloc)
	v28 := rand.Float64()

	return []types.Metrics{
		{ID: "Alloc", MType: types.Gauge, Value: &v1},
		{ID: "BuckHashSys", MType: types.Gauge, Value: &v2},
		{ID: "Frees", MType: types.Gauge, Value: &v3},
		{ID: "GCCPUFraction", MType: types.Gauge, Value: &v4},
		{ID: "GCSys", MType: types.Gauge, Value: &v5},
		{ID: "HeapAlloc", MType: types.Gauge, Value: &v6},
		{ID: "HeapIdle", MType: types.Gauge, Value: &v7},
		{ID: "HeapInuse", MType: types.Gauge, Value: &v8},
		{ID: "HeapObjects", MType: types.Gauge, Value: &v9},
		{ID: "HeapReleased", MType: types.Gauge, Value: &v10},
		{ID: "HeapSys", MType: types.Gauge, Value: &v11},
		{ID: "LastGC", MType: types.Gauge, Value: &v12},
		{ID: "Lookups", MType: types.Gauge, Value: &v13},
		{ID: "MCacheInuse", MType: types.Gauge, Value: &v14},
		{ID: "MCacheSys", MType: types.Gauge, Value: &v15},
		{ID: "MSpanInuse", MType: types.Gauge, Value: &v16},
		{ID: "MSpanSys", MType: types.Gauge, Value: &v17},
		{ID: "Mallocs", MType: types.Gauge, Value: &v18},
		{ID: "NextGC", MType: types.Gauge, Value: &v19},
		{ID: "NumForcedGC", MType: types.Gauge, Value: &v20},
		{ID: "NumGC", MType: types.Gauge, Value: &v21},
		{ID: "OtherSys", MType: types.Gauge, Value: &v22},
		{ID: "PauseTotalNs", MType: types.Gauge, Value: &v23},
		{ID: "StackInuse", MType: types.Gauge, Value: &v24},
		{ID: "StackSys", MType: types.Gauge, Value: &v25},
		{ID: "Sys", MType: types.Gauge, Value: &v26},
		{ID: "TotalAlloc", MType: types.Gauge, Value: &v27},
		{ID: "RandomValue", MType: types.Gauge, Value: &v28},
	}
}

func getRuntimeCounterMetrics() []types.Metrics {
	pollCount := int64(1)

	return []types.Metrics{
		{
			ID:    "PollCount",
			MType: types.Counter,
			Delta: &pollCount,
		},
	}
}

func getGoputilMetrics(ctx context.Context) []types.Metrics {
	var result []types.Metrics

	if vmStat, err := mem.VirtualMemory(); err == nil {
		total := float64(vmStat.Total)
		free := float64(vmStat.Free)

		result = append(result, types.Metrics{
			ID:    "TotalMemory",
			MType: types.Gauge,
			Value: &total,
		})
		result = append(result, types.Metrics{
			ID:    "FreeMemory",
			MType: types.Gauge,
			Value: &free,
		})
	}

	if cpuPercents, err := cpu.PercentWithContext(ctx, 0, true); err == nil {
		for i, percent := range cpuPercents {
			p := percent
			id := "CPUutilization" + strconv.Itoa(i)
			result = append(result, types.Metrics{
				ID:    id,
				MType: types.Gauge,
				Value: &p,
			})
		}
	}

	return result
}

func generatorMetrics(
	ctx context.Context,
	input []types.Metrics,
) chan types.Metrics {
	out := make(chan types.Metrics, 100)

	go func() {
		defer close(out)

		select {
		case <-ctx.Done():
			return
		default:
		}

		for _, item := range input {
			out <- item
		}
	}()

	return out
}

func fanInMetrics(ctx context.Context, resultChs ...chan types.Metrics) chan types.Metrics {
	finalCh := make(chan types.Metrics)

	var wg sync.WaitGroup

	for _, ch := range resultChs {
		chClosure := ch
		wg.Add(1)

		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case data, ok := <-chClosure:
					if !ok {
						return
					}
					finalCh <- data
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}
