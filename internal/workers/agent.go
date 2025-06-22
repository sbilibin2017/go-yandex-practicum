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
)

// Updater defines an interface for updating metrics.
// Implementers should provide the Updates method that sends metrics data somewhere.
type Updater interface {
	Updates(ctx context.Context, metrics []*types.Metrics) error
}

// Result represents the outcome of a metrics update operation,
// containing the data updated and any error that occurred.
type result struct {
	Data []*types.Metrics
	Err  error
}

// AgentWorkerConfig holds configuration options for the agent worker.
type agentWorkerOptions struct {
	pollInterval   int
	reportInterval int
	batchSize      int
	rateLimit      int
	updater        Updater
}

// AgentWorkerOption represents a functional option for configuring AgentWorkerConfig.
type AgentWorkerOption func(*agentWorkerOptions)

// WithPollInterval sets the polling interval (in seconds) for metrics collection.
func WithPollInterval(interval int) AgentWorkerOption {
	return func(cfg *agentWorkerOptions) {
		cfg.pollInterval = interval
	}
}

// WithReportInterval sets the reporting interval (in seconds) for sending metrics updates.
func WithReportInterval(interval int) AgentWorkerOption {
	return func(cfg *agentWorkerOptions) {
		cfg.reportInterval = interval
	}
}

// WithBatchSize sets the maximum number of metrics to batch before sending an update.
func WithBatchSize(size int) AgentWorkerOption {
	return func(cfg *agentWorkerOptions) {
		cfg.batchSize = size
	}
}

// WithRateLimit sets the maximum number of concurrent workers reporting metrics.
func WithRateLimit(limit int) AgentWorkerOption {
	return func(cfg *agentWorkerOptions) {
		cfg.rateLimit = limit
	}
}

// WithUpdater sets the Updater implementation that handles sending metrics updates.
func WithUpdater(updater Updater) AgentWorkerOption {
	return func(cfg *agentWorkerOptions) {
		cfg.updater = updater
	}
}

// NewAgentWorker creates and returns a worker function that collects and reports metrics
// according to the provided configuration options.
func NewAgentWorker(opts ...AgentWorkerOption) func(ctx context.Context) error {
	cfg := &agentWorkerOptions{
		// Set default values here if needed
		pollInterval:   2,
		reportInterval: 10,
		batchSize:      10,
		rateLimit:      runtime.NumCPU(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return func(ctx context.Context) error {
		return startAgent(
			ctx,
			cfg.updater,
			cfg.pollInterval,
			cfg.reportInterval,
			cfg.batchSize,
			cfg.rateLimit,
		)
	}
}

// startAgent coordinates the polling and reporting of metrics,
// starting necessary goroutines and managing their lifecycle.
func startAgent(
	ctx context.Context,
	updater Updater,
	pollInterval int,
	reportInterval int,
	batchSize int,
	rateLimit int,
) error {
	pollCh := startMetricsPolling(ctx, pollInterval)
	reportCh := startMetricsReporting(ctx, reportInterval, updater, pollCh, batchSize, rateLimit)
	return logResults(ctx, reportCh)
}

// startMetricsPolling begins periodic metrics collection.
// Returns a channel that streams collected metrics.
func startMetricsPolling(ctx context.Context, pollInterval int) <-chan *types.Metrics {
	getRuntimeGaugeMetrics := func() []*types.Metrics {
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

		return []*types.Metrics{
			{ID: "Alloc", Type: types.Gauge, Value: &v1},
			{ID: "BuckHashSys", Type: types.Gauge, Value: &v2},
			{ID: "Frees", Type: types.Gauge, Value: &v3},
			{ID: "GCCPUFraction", Type: types.Gauge, Value: &v4},
			{ID: "GCSys", Type: types.Gauge, Value: &v5},
			{ID: "HeapAlloc", Type: types.Gauge, Value: &v6},
			{ID: "HeapIdle", Type: types.Gauge, Value: &v7},
			{ID: "HeapInuse", Type: types.Gauge, Value: &v8},
			{ID: "HeapObjects", Type: types.Gauge, Value: &v9},
			{ID: "HeapReleased", Type: types.Gauge, Value: &v10},
			{ID: "HeapSys", Type: types.Gauge, Value: &v11},
			{ID: "LastGC", Type: types.Gauge, Value: &v12},
			{ID: "Lookups", Type: types.Gauge, Value: &v13},
			{ID: "MCacheInuse", Type: types.Gauge, Value: &v14},
			{ID: "MCacheSys", Type: types.Gauge, Value: &v15},
			{ID: "MSpanInuse", Type: types.Gauge, Value: &v16},
			{ID: "MSpanSys", Type: types.Gauge, Value: &v17},
			{ID: "Mallocs", Type: types.Gauge, Value: &v18},
			{ID: "NextGC", Type: types.Gauge, Value: &v19},
			{ID: "NumForcedGC", Type: types.Gauge, Value: &v20},
			{ID: "NumGC", Type: types.Gauge, Value: &v21},
			{ID: "OtherSys", Type: types.Gauge, Value: &v22},
			{ID: "PauseTotalNs", Type: types.Gauge, Value: &v23},
			{ID: "StackInuse", Type: types.Gauge, Value: &v24},
			{ID: "StackSys", Type: types.Gauge, Value: &v25},
			{ID: "Sys", Type: types.Gauge, Value: &v26},
			{ID: "TotalAlloc", Type: types.Gauge, Value: &v27},
			{ID: "RandomValue", Type: types.Gauge, Value: &v28},
		}
	}

	getRuntimeCounterMetrics := func() []*types.Metrics {
		pollCount := int64(1)
		return []*types.Metrics{
			{ID: "PollCount", Type: types.Counter, Delta: &pollCount},
		}
	}

	getGoputilMetrics := func(ctx context.Context) []*types.Metrics {
		var result []*types.Metrics

		if vmStat, err := mem.VirtualMemory(); err == nil {
			total := float64(vmStat.Total)
			free := float64(vmStat.Free)

			result = append(result, &types.Metrics{ID: "TotalMemory", Type: types.Gauge, Value: &total})
			result = append(result, &types.Metrics{ID: "FreeMemory", Type: types.Gauge, Value: &free})
		}

		if cpuPercents, err := cpu.PercentWithContext(ctx, 0, true); err == nil {
			for i, percent := range cpuPercents {
				p := percent
				id := "CPUutilization" + strconv.Itoa(i)
				result = append(result, &types.Metrics{ID: id, Type: types.Gauge, Value: &p})
			}
		}

		return result
	}

	generatorMetrics := func(input []*types.Metrics) chan *types.Metrics {
		out := make(chan *types.Metrics, 100)

		go func() {
			defer close(out)
			for _, item := range input {
				out <- item
			}
		}()

		return out
	}

	fanInMetrics := func(ctx context.Context, channels ...chan *types.Metrics) chan *types.Metrics {
		finalCh := make(chan *types.Metrics)
		var wg sync.WaitGroup

		for _, ch := range channels {
			chCopy := ch
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case metric, ok := <-chCopy:
						if !ok {
							return
						}
						finalCh <- metric
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

	out := make(chan *types.Metrics, 100)

	go func() {
		defer close(out)

		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runtimeCh := generatorMetrics(getRuntimeGaugeMetrics())
				counterCh := generatorMetrics(getRuntimeCounterMetrics())
				gopsutilCh := generatorMetrics(getGoputilMetrics(ctx))

				for metric := range fanInMetrics(ctx, runtimeCh, counterCh, gopsutilCh) {
					out <- metric
				}
			}
		}
	}()

	return out
}

// startMetricsReporting starts the reporting workers that batch and send metrics updates.
// It returns a channel of Result indicating success or failure of update operations.
func startMetricsReporting(
	ctx context.Context,
	reportInterval int,
	updater Updater,
	in <-chan *types.Metrics,
	batchSize int,
	rateLimit int,
) <-chan result {
	logger.Log.Debug("startMetricsReporting: starting")

	resultsCh := make(chan result, 100)
	wg := &sync.WaitGroup{}

	worker := func(workerID int, wg *sync.WaitGroup, jobsCh <-chan *types.Metrics) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-jobsCh:
				if !ok {
					return
				}
				batch := []*types.Metrics{m}

			collectLoop:
				for len(batch) < batchSize {
					select {
					case m, ok := <-jobsCh:
						if !ok {
							break collectLoop
						}
						batch = append(batch, m)
					default:
						break collectLoop
					}
				}

				err := updater.Updates(ctx, batch)
				if err != nil {
					logger.Log.Error(err)
				}

				resultsCh <- result{Data: batch, Err: err}
			}
		}
	}

	go func() {
		defer close(resultsCh)

		ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				wg.Add(rateLimit)
				jobsCh := make(chan *types.Metrics, 100)

				for i := 0; i < rateLimit; i++ {
					go worker(i, wg, jobsCh)
				}

			feedLoop:
				for {
					select {
					case <-ctx.Done():
						break feedLoop
					case m, ok := <-in:
						if !ok {
							break feedLoop
						}
						jobsCh <- m
					default:
						break feedLoop
					}
				}

				close(jobsCh)
				wg.Wait()
			}
		}
	}()

	return resultsCh
}

// logResults consumes the results channel and logs errors if any occur.
// It returns context error when the context is done.
func logResults(ctx context.Context, results <-chan result) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case res, ok := <-results:
				if !ok {
					return
				}
				if res.Err != nil {
					logger.Log.Errorw("logResults: error updating metrics", "error", res.Err, "dataCount", len(res.Data))
				}
			}
		}
	}()
	return ctx.Err()
}
