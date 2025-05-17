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

type result struct {
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

func NewMetricAgentWorker(
	ctx context.Context,
	facade MetricFacade,
	sema Semaphore,
	flagPollInterval int,
	flagReportInterval int,
	flagNumWorkers int,
	flagBatchSize int,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return startMetricAgent(
			ctx,
			facade,
			sema,
			flagPollInterval,
			flagReportInterval,
			flagNumWorkers,
			flagBatchSize,
		)
	}
}

func startMetricAgent(
	ctx context.Context,
	facade MetricFacade,
	sema Semaphore,
	flagPollInterval int,
	flagReportInterval int,
	flagNumWorkers int,
	flagBatchSize int,
) error {
	pollInterval := time.Duration(flagPollInterval) * time.Second
	reportInterval := time.Duration(flagPollInterval) * time.Second
	runtimeCh := generatorRuntimeGaugeMetrics(ctx)
	counterCh := generatorRuntimeCounterMetrics(ctx)
	gopsutilCh := generatorGoputilMetrics(ctx)
	allMetricsCh := metricsFanIn(ctx, runtimeCh, counterCh, gopsutilCh)
	polledCh := pollMetrics(ctx, pollInterval, allMetricsCh)
	reportCh := reportMetrics(ctx, reportInterval, polledCh)
	resultChs := metricsFanOut(ctx, facade, sema, flagNumWorkers, flagBatchSize, reportCh)
	errCh := processResults(ctx, resultChs)
	return waitForErrors(ctx, errCh)
}

func waitForErrors(ctx context.Context, errCh <-chan error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err, ok := <-errCh:
			if !ok {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}
}

func generatorRuntimeGaugeMetrics(ctx context.Context) chan types.Metrics {
	out := make(chan types.Metrics)

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var memStats runtime.MemStats
				runtime.ReadMemStats(&memStats)

				v1 := float64(memStats.Alloc)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "Alloc", Type: types.GaugeMetricType},
					Value:    &v1,
				}

				v2 := float64(memStats.BuckHashSys)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "BuckHashSys", Type: types.GaugeMetricType},
					Value:    &v2,
				}

				v3 := float64(memStats.Frees)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "Frees", Type: types.GaugeMetricType},
					Value:    &v3,
				}

				v4 := memStats.GCCPUFraction
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "GCCPUFraction", Type: types.GaugeMetricType},
					Value:    &v4,
				}

				v5 := float64(memStats.GCSys)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "GCSys", Type: types.GaugeMetricType},
					Value:    &v5,
				}

				v6 := float64(memStats.HeapAlloc)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "HeapAlloc", Type: types.GaugeMetricType},
					Value:    &v6,
				}

				v7 := float64(memStats.HeapIdle)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "HeapIdle", Type: types.GaugeMetricType},
					Value:    &v7,
				}

				v8 := float64(memStats.HeapInuse)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "HeapInuse", Type: types.GaugeMetricType},
					Value:    &v8,
				}

				v9 := float64(memStats.HeapObjects)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "HeapObjects", Type: types.GaugeMetricType},
					Value:    &v9,
				}

				v10 := float64(memStats.HeapReleased)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "HeapReleased", Type: types.GaugeMetricType},
					Value:    &v10,
				}

				v11 := float64(memStats.HeapSys)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "HeapSys", Type: types.GaugeMetricType},
					Value:    &v11,
				}

				v12 := float64(memStats.LastGC)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "LastGC", Type: types.GaugeMetricType},
					Value:    &v12,
				}

				v13 := float64(memStats.Lookups)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "Lookups", Type: types.GaugeMetricType},
					Value:    &v13,
				}

				v14 := float64(memStats.MCacheInuse)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "MCacheInuse", Type: types.GaugeMetricType},
					Value:    &v14,
				}

				v15 := float64(memStats.MCacheSys)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "MCacheSys", Type: types.GaugeMetricType},
					Value:    &v15,
				}

				v16 := float64(memStats.MSpanInuse)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "MSpanInuse", Type: types.GaugeMetricType},
					Value:    &v16,
				}

				v17 := float64(memStats.MSpanSys)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "MSpanSys", Type: types.GaugeMetricType},
					Value:    &v17,
				}

				v18 := float64(memStats.Mallocs)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "Mallocs", Type: types.GaugeMetricType},
					Value:    &v18,
				}

				v19 := float64(memStats.NextGC)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "NextGC", Type: types.GaugeMetricType},
					Value:    &v19,
				}

				v20 := float64(memStats.NumForcedGC)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "NumForcedGC", Type: types.GaugeMetricType},
					Value:    &v20,
				}

				v21 := float64(memStats.NumGC)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "NumGC", Type: types.GaugeMetricType},
					Value:    &v21,
				}

				v22 := float64(memStats.OtherSys)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "OtherSys", Type: types.GaugeMetricType},
					Value:    &v22,
				}

				v23 := float64(memStats.PauseTotalNs)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "PauseTotalNs", Type: types.GaugeMetricType},
					Value:    &v23,
				}

				v24 := float64(memStats.StackInuse)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "StackInuse", Type: types.GaugeMetricType},
					Value:    &v24,
				}

				v25 := float64(memStats.StackSys)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "StackSys", Type: types.GaugeMetricType},
					Value:    &v25,
				}

				v26 := float64(memStats.Sys)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "Sys", Type: types.GaugeMetricType},
					Value:    &v26,
				}

				v27 := float64(memStats.TotalAlloc)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "TotalAlloc", Type: types.GaugeMetricType},
					Value:    &v27,
				}

				v28 := rand.Float64()
				out <- types.Metrics{
					MetricID: types.MetricID{ID: "RandomValue", Type: types.GaugeMetricType},
					Value:    &v28,
				}
			}
		}
	}()

	return out
}

func generatorRuntimeCounterMetrics(ctx context.Context) chan types.Metrics {
	out := make(chan types.Metrics, 2)

	go func() {
		defer close(out)
		v := int64(1)
		select {
		case <-ctx.Done():
			return
		case out <- types.Metrics{
			MetricID: types.MetricID{ID: "PollCount", Type: types.CounterMetricType},
			Delta:    &v,
		}:
		}
	}()

	return out
}

func generatorGoputilMetrics(ctx context.Context) chan types.Metrics {
	out := make(chan types.Metrics, 10)

	go func() {
		defer close(out)

		select {
		case <-ctx.Done():
			return
		default:
		}

		if vmStat, err := mem.VirtualMemory(); err == nil {
			total := float64(vmStat.Total)
			free := float64(vmStat.Free)
			out <- types.Metrics{
				MetricID: types.MetricID{ID: "TotalMemory", Type: types.GaugeMetricType},
				Value:    &total,
			}
			out <- types.Metrics{
				MetricID: types.MetricID{ID: "FreeMemory", Type: types.GaugeMetricType},
				Value:    &free,
			}
		}

		if cpuPercents, err := cpu.PercentWithContext(ctx, 0, true); err == nil {
			for i, percent := range cpuPercents {
				p := percent
				id := "CPUutilization" + strconv.Itoa(i)
				out <- types.Metrics{
					MetricID: types.MetricID{ID: id, Type: types.GaugeMetricType},
					Value:    &p,
				}
			}
		}
	}()

	return out
}

func metricsFanIn(ctx context.Context, resultChs ...chan types.Metrics) chan types.Metrics {
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

func metricsHandler(
	ctx context.Context,
	facade MetricFacade,
	inputCh chan types.Metrics,
	batchSize int,
) chan result {
	resultCh := make(chan result)

	go func() {
		defer close(resultCh)

		batch := make([]types.Metrics, 0, batchSize)

		for {
			select {
			case <-ctx.Done():
				return
			case metric, ok := <-inputCh:
				if !ok {
					if len(batch) > 0 {
						err := facade.Updates(ctx, batch)
						for _, m := range batch {
							resultCh <- result{data: m, err: err}
						}
					}
					return
				}

				batch = append(batch, metric)

				if len(batch) >= batchSize {
					err := facade.Updates(ctx, batch)
					for _, m := range batch {
						resultCh <- result{data: m, err: err}
					}
					batch = batch[:0]
				}
			}
		}
	}()

	return resultCh
}

func metricsFanOut(
	ctx context.Context,
	facade MetricFacade,
	sema Semaphore,
	numWorkers int,
	batchSize int,
	inputCh chan types.Metrics,
) []chan result {
	channels := make([]chan result, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()
			if err := sema.Acquire(ctx, 1); err != nil {
				logger.Log.Error("failed to acquire semaphore", zap.Int("worker", idx), zap.Error(err))
				return
			}
			defer sema.Release(1)
			channels[idx] = metricsHandler(ctx, facade, inputCh, batchSize)
		}(i)
	}

	wg.Wait()

	return channels
}

func pollMetrics(ctx context.Context, pollInterval time.Duration, metricsInputCh <-chan types.Metrics) chan types.Metrics {
	batchedInputCh := make(chan types.Metrics)

	go func() {
		ticker := time.NewTicker(pollInterval)
		defer func() {
			ticker.Stop()
			close(batchedInputCh)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m, ok := <-metricsInputCh
				if !ok {
					return
				}
				batchedInputCh <- m
			}
		}
	}()

	return batchedInputCh
}

func reportMetrics(
	ctx context.Context,
	pollInterval time.Duration,
	metricsInputCh <-chan types.Metrics,
) chan types.Metrics {
	batchedInputCh := make(chan types.Metrics)

	go func() {
		ticker := time.NewTicker(pollInterval)
		defer func() {
			ticker.Stop()
			close(batchedInputCh)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			loop:
				for {
					select {
					case m, ok := <-metricsInputCh:
						if !ok {
							return
						}
						batchedInputCh <- m
					default:
						break loop
					}
				}
			}
		}
	}()

	return batchedInputCh
}

func processResults(ctx context.Context, resultChs []chan result) chan error {
	errCh := make(chan error)

	var wg sync.WaitGroup

	for _, ch := range resultChs {
		ch := ch
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case res, ok := <-ch:
					if !ok {
						return
					}
					if res.err != nil {
						logger.Log.Error("ошибка при отправке метрики",
							zap.String("metric_id", res.data.MetricID.ID),
							zap.String("metric_type", string(res.data.Type)),
							zap.Error(res.err),
						)
						errCh <- res.err
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	return errCh
}
