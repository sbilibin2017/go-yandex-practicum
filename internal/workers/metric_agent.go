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
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type Result struct {
	data types.Metrics
	err  error
}

type MetricFacade interface {
	Updates(ctx context.Context, metrics []types.Metrics) error
}

type MetricAgentWorker struct {
	metricFacade   MetricFacade
	pollInterval   int
	reportInterval int
	numWorkers     int
}

func NewMetricAgentWorker(
	metricFacade MetricFacade,
	pollInterval int,
	reportInterval int,
	numWorkers int,
) *MetricAgentWorker {
	return &MetricAgentWorker{
		metricFacade:   metricFacade,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		numWorkers:     numWorkers,
	}
}

func (w *MetricAgentWorker) Start(ctx context.Context) error {
	pollInterval := w.pollInterval
	reportInterval := w.reportInterval

	if pollInterval == 0 {
		pollInterval = 1
	}
	if reportInterval == 0 {
		reportInterval = 1
	}

	pollTicker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer reportTicker.Stop()

	generatorCh := generator(ctx, pollTicker)

	runtimeStageCh := produceRuntimeGaugeMetricsStage(ctx, generatorCh)
	runtimeCounterCh := produceRuntimeCounterMetricsStage(ctx, runtimeStageCh)
	gopsutilCh := produceGopsutilGaugeMetricsStage(ctx, generatorCh)

	metricsCh := metricsFanOut(ctx, runtimeCounterCh, gopsutilCh)

	resultChs := metricsWorkerPool(ctx, w.metricFacade, metricsCh, reportTicker, w.numWorkers)

	finalResultCh := fanInResults(ctx, resultChs...)

	errCh := make(chan error, 1)
	handleResults(ctx, finalResultCh, errCh)

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func generator(ctx context.Context, pollTicker *time.Ticker) <-chan types.Metrics {
	out := make(chan types.Metrics)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case <-pollTicker.C:
				out <- types.Metrics{}
			}
		}
	}()
	return out
}

func produceRuntimeGaugeMetricsStage(ctx context.Context, in <-chan types.Metrics) <-chan types.Metrics {
	out := make(chan types.Metrics)
	go func() {
		defer close(out)
		for range in {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			values := []struct {
				id    string
				value float64
			}{
				{"Alloc", float64(memStats.Alloc)},
				{"BuckHashSys", float64(memStats.BuckHashSys)},
				{"Frees", float64(memStats.Frees)},
				{"GCCPUFraction", memStats.GCCPUFraction},
				{"GCSys", float64(memStats.GCSys)},
				{"HeapAlloc", float64(memStats.HeapAlloc)},
				{"HeapIdle", float64(memStats.HeapIdle)},
				{"HeapInuse", float64(memStats.HeapInuse)},
				{"HeapObjects", float64(memStats.HeapObjects)},
				{"HeapReleased", float64(memStats.HeapReleased)},
				{"HeapSys", float64(memStats.HeapSys)},
				{"LastGC", float64(memStats.LastGC)},
				{"Lookups", float64(memStats.Lookups)},
				{"MCacheInuse", float64(memStats.MCacheInuse)},
				{"MCacheSys", float64(memStats.MCacheSys)},
				{"MSpanInuse", float64(memStats.MSpanInuse)},
				{"MSpanSys", float64(memStats.MSpanSys)},
				{"Mallocs", float64(memStats.Mallocs)},
				{"NextGC", float64(memStats.NextGC)},
				{"NumForcedGC", float64(memStats.NumForcedGC)},
				{"NumGC", float64(memStats.NumGC)},
				{"OtherSys", float64(memStats.OtherSys)},
				{"PauseTotalNs", float64(memStats.PauseTotalNs)},
				{"StackInuse", float64(memStats.StackInuse)},
				{"StackSys", float64(memStats.StackSys)},
				{"Sys", float64(memStats.Sys)},
				{"TotalAlloc", float64(memStats.TotalAlloc)},
				{"RandomValue", rand.Float64()},
			}

			for _, v := range values {
				select {
				case <-ctx.Done():
					return
				case out <- types.Metrics{MetricID: types.MetricID{ID: v.id, Type: types.GaugeMetricType}, Value: &v.value}:
				}
			}
		}
	}()
	return out
}

func produceRuntimeCounterMetricsStage(ctx context.Context, in <-chan types.Metrics) <-chan types.Metrics {
	out := make(chan types.Metrics)
	go func() {
		defer close(out)
		deltas := []struct {
			id    string
			delta int64
		}{
			{"PollCount", 1},
		}
		for _, d := range deltas {
			select {
			case <-ctx.Done():
				return
			case out <- types.Metrics{MetricID: types.MetricID{ID: d.id, Type: types.CounterMetricType}, Delta: &d.delta}:
			}
		}
	}()
	return out
}

func produceGopsutilGaugeMetricsStage(ctx context.Context, in <-chan types.Metrics) <-chan types.Metrics {
	out := make(chan types.Metrics, 10)
	go func() {
		defer close(out)
		for range in {
			if vmStat, err := mem.VirtualMemory(); err == nil {
				total := float64(vmStat.Total)
				free := float64(vmStat.Free)

				select {
				case <-ctx.Done():
					return
				case out <- types.Metrics{MetricID: types.MetricID{ID: "TotalMemory", Type: types.GaugeMetricType}, Value: &total}:
				}

				select {
				case <-ctx.Done():
					return
				case out <- types.Metrics{MetricID: types.MetricID{ID: "FreeMemory", Type: types.GaugeMetricType}, Value: &free}:
				}
			}
			if cpuPercents, err := cpu.Percent(0, true); err == nil {
				for i, percent := range cpuPercents {
					id := "CPUutilization" + strconv.Itoa(i)
					p := percent

					select {
					case <-ctx.Done():
						return
					case out <- types.Metrics{MetricID: types.MetricID{ID: id, Type: types.GaugeMetricType}, Value: &p}:
					}
				}
			}
		}
	}()
	return out
}

func metricsFanOut(ctx context.Context, chans ...<-chan types.Metrics) <-chan types.Metrics {
	var wg sync.WaitGroup
	out := make(chan types.Metrics)

	output := func(c <-chan types.Metrics) {
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
				case <-ctx.Done():
					return
				case out <- m:
				}
			}
		}
	}

	wg.Add(len(chans))
	for _, c := range chans {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func metricsWorker(
	ctx context.Context,
	handler MetricFacade,
	taskCh <-chan types.Metrics,
	reportTicker *time.Ticker,
	sem *semaphore.Weighted,
) <-chan Result {
	resultCh := make(chan Result, 1000)
	metricsBatch := make([]types.Metrics, 0, 1000)

	go func() {
		defer close(resultCh)

		if err := sem.Acquire(ctx, 1); err != nil {
			return
		}
		defer sem.Release(1)

		for {
			select {
			case <-ctx.Done():
				return

			case metric, ok := <-taskCh:
				if !ok {
					if len(metricsBatch) > 0 {
						err := handler.Updates(ctx, metricsBatch)
						for _, metric := range metricsBatch {
							resultCh <- Result{data: metric, err: err}
						}
					}
					return
				}
				metricsBatch = append(metricsBatch, metric)

			case <-reportTicker.C:
				if len(metricsBatch) == 0 {
					continue
				}
				err := handler.Updates(ctx, metricsBatch)
				for _, metric := range metricsBatch {
					resultCh <- Result{data: metric, err: err}
				}
				metricsBatch = metricsBatch[:0]
			}
		}
	}()

	return resultCh
}

func metricsWorkerPool(
	ctx context.Context,
	handler MetricFacade,
	taskCh <-chan types.Metrics,
	reportTicker *time.Ticker,
	numWorkers int,
) []<-chan Result {
	sem := semaphore.NewWeighted(int64(numWorkers))
	resultChs := make([]<-chan Result, numWorkers)
	for i := 0; i < numWorkers; i++ {
		resultChs[i] = metricsWorker(ctx, handler, taskCh, reportTicker, sem)
	}
	return resultChs
}

func fanInResults(ctx context.Context, resultChs ...<-chan Result) <-chan Result {
	finalCh := make(chan Result)
	var wg sync.WaitGroup

	for _, ch := range resultChs {
		wg.Add(1)
		go func(c <-chan Result) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case res, ok := <-c:
					if !ok {
						return
					}
					finalCh <- res
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}

func handleResults(ctx context.Context, resultCh <-chan Result, errCh chan<- error) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case res, ok := <-resultCh:
				if !ok {
					return
				}
				if res.err != nil {
					logger.Log.Error("Ошибка отправки метрики",
						zap.String("metric", res.data.MetricID.ID),
						zap.Error(res.err),
					)
					// Передаём первую ошибку
					select {
					case errCh <- res.err:
					default:
					}
				} else {
					logger.Log.Info("Метрика успешно отправлена",
						zap.String("metric", res.data.MetricID.ID),
					)
				}
			}
		}
	}()
}
