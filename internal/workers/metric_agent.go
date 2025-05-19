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

type MetricFacade interface {
	Updates(ctx context.Context, metrics []types.Metrics) error
}

type result struct {
	Data []types.Metrics
	Err  error
}

func NewMetricAgentWorker(
	facade MetricFacade,
	pollInterval int,
	reportInterval int,
	rateLimit int,
	batchSize int,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return startMetricAgent(
			ctx,
			facade,
			pollInterval,
			reportInterval,
			rateLimit,
			batchSize,
		)
	}
}

func startMetricAgent(
	ctx context.Context,
	facade MetricFacade,
	pollInterval int,
	reportInterval int,
	rateLimit int,
	batchSize int,
) error {
	pollTicker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer reportTicker.Stop()

	var metricsFanInCh chan types.Metrics

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-pollTicker.C:
			runtimeCh := generatorMetrics(ctx, getRuntimeGaugeMetrics)
			counterCh := generatorMetrics(ctx, getRuntimeCounterMetrics)
			gopsutilCh := generatorMetrics(ctx, getGoputilMetrics)
			metricsFanInCh = fanInMetrics(ctx, runtimeCh, counterCh, gopsutilCh)

		case <-reportTicker.C:
			results := workerPoolMetricsUpdate(
				ctx,
				facade,
				workerMetricsUpdate,
				metricsFanInCh,
				rateLimit,
				batchSize,
			)
			processMetricResults(results)
			metricsFanInCh = nil
		}
	}
}

func getRuntimeGaugeMetrics(ctx context.Context) []types.Metrics {
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
		{MetricID: types.MetricID{ID: "Alloc", Type: types.GaugeMetricType}, Value: &v1},
		{MetricID: types.MetricID{ID: "BuckHashSys", Type: types.GaugeMetricType}, Value: &v2},
		{MetricID: types.MetricID{ID: "Frees", Type: types.GaugeMetricType}, Value: &v3},
		{MetricID: types.MetricID{ID: "GCCPUFraction", Type: types.GaugeMetricType}, Value: &v4},
		{MetricID: types.MetricID{ID: "GCSys", Type: types.GaugeMetricType}, Value: &v5},
		{MetricID: types.MetricID{ID: "HeapAlloc", Type: types.GaugeMetricType}, Value: &v6},
		{MetricID: types.MetricID{ID: "HeapIdle", Type: types.GaugeMetricType}, Value: &v7},
		{MetricID: types.MetricID{ID: "HeapInuse", Type: types.GaugeMetricType}, Value: &v8},
		{MetricID: types.MetricID{ID: "HeapObjects", Type: types.GaugeMetricType}, Value: &v9},
		{MetricID: types.MetricID{ID: "HeapReleased", Type: types.GaugeMetricType}, Value: &v10},
		{MetricID: types.MetricID{ID: "HeapSys", Type: types.GaugeMetricType}, Value: &v11},
		{MetricID: types.MetricID{ID: "LastGC", Type: types.GaugeMetricType}, Value: &v12},
		{MetricID: types.MetricID{ID: "Lookups", Type: types.GaugeMetricType}, Value: &v13},
		{MetricID: types.MetricID{ID: "MCacheInuse", Type: types.GaugeMetricType}, Value: &v14},
		{MetricID: types.MetricID{ID: "MCacheSys", Type: types.GaugeMetricType}, Value: &v15},
		{MetricID: types.MetricID{ID: "MSpanInuse", Type: types.GaugeMetricType}, Value: &v16},
		{MetricID: types.MetricID{ID: "MSpanSys", Type: types.GaugeMetricType}, Value: &v17},
		{MetricID: types.MetricID{ID: "Mallocs", Type: types.GaugeMetricType}, Value: &v18},
		{MetricID: types.MetricID{ID: "NextGC", Type: types.GaugeMetricType}, Value: &v19},
		{MetricID: types.MetricID{ID: "NumForcedGC", Type: types.GaugeMetricType}, Value: &v20},
		{MetricID: types.MetricID{ID: "NumGC", Type: types.GaugeMetricType}, Value: &v21},
		{MetricID: types.MetricID{ID: "OtherSys", Type: types.GaugeMetricType}, Value: &v22},
		{MetricID: types.MetricID{ID: "PauseTotalNs", Type: types.GaugeMetricType}, Value: &v23},
		{MetricID: types.MetricID{ID: "StackInuse", Type: types.GaugeMetricType}, Value: &v24},
		{MetricID: types.MetricID{ID: "StackSys", Type: types.GaugeMetricType}, Value: &v25},
		{MetricID: types.MetricID{ID: "Sys", Type: types.GaugeMetricType}, Value: &v26},
		{MetricID: types.MetricID{ID: "TotalAlloc", Type: types.GaugeMetricType}, Value: &v27},
		{MetricID: types.MetricID{ID: "RandomValue", Type: types.GaugeMetricType}, Value: &v28},
	}
}

func getRuntimeCounterMetrics(ctx context.Context) []types.Metrics {
	pollCount := int64(1)

	return []types.Metrics{
		{
			MetricID: types.MetricID{ID: "PollCount", Type: types.CounterMetricType},
			Delta:    &pollCount,
		},
	}
}

func getGoputilMetrics(ctx context.Context) []types.Metrics {
	var result []types.Metrics

	if vmStat, err := mem.VirtualMemory(); err == nil {
		total := float64(vmStat.Total)
		free := float64(vmStat.Free)

		result = append(result, types.Metrics{
			MetricID: types.MetricID{ID: "TotalMemory", Type: types.GaugeMetricType},
			Value:    &total,
		})
		result = append(result, types.Metrics{
			MetricID: types.MetricID{ID: "FreeMemory", Type: types.GaugeMetricType},
			Value:    &free,
		})
	}

	if cpuPercents, err := cpu.PercentWithContext(ctx, 0, true); err == nil {
		for i, percent := range cpuPercents {
			p := percent
			id := "CPUutilization" + strconv.Itoa(i)
			result = append(result, types.Metrics{
				MetricID: types.MetricID{ID: id, Type: types.GaugeMetricType},
				Value:    &p,
			})
		}
	}

	return result
}

func generatorMetrics(
	ctx context.Context,
	inputFunc func(ctx context.Context) []types.Metrics,
) chan types.Metrics {
	out := make(chan types.Metrics, 100)

	go func() {
		defer close(out)

		select {
		case <-ctx.Done():
			return
		default:
		}

		for _, item := range inputFunc(ctx) {
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
					select {
					case finalCh <- data:
					case <-ctx.Done():
						return
					}
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

func workerMetricsUpdate(
	ctx context.Context,
	facade MetricFacade,
	job chan types.Metrics,
	results chan result,
	batchSize int,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-job:
			if !ok {
				return
			}

			batch := []types.Metrics{m}

		collectLoop:
			for len(batch) < batchSize {
				select {
				case m, ok := <-job:
					if !ok {
						break collectLoop
					}
					batch = append(batch, m)
				default:
					break collectLoop
				}
			}

			err := facade.Updates(ctx, batch)

			results <- result{
				Data: batch,
				Err:  err,
			}
		}
	}
}

func workerPoolMetricsUpdate(
	ctx context.Context,
	facade MetricFacade,
	worker func(ctx context.Context, facade MetricFacade, job chan types.Metrics, results chan result, batchSize int),
	jobs chan types.Metrics,
	numWorkers int,
	batchSize int,
) chan result {
	results := make(chan result)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			worker(ctx, facade, jobs, results, batchSize)
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func processMetricResults(results <-chan result) {
	for res := range results {
		if res.Err != nil {
			logger.Log.Error("worker pool task error", zap.Error(res.Err), zap.Any("data", res.Data))
		} else {
			logger.Log.Info("worker pool task success", zap.Any("data", res.Data))
		}
	}
}
