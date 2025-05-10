package workers

import (
	"context"
	"math/rand"
	"runtime"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricFacade interface {
	Update(ctx context.Context, metric map[string]any) error
}

func StartMetricAgent(
	ctx context.Context,
	facade MetricFacade,
	pollTicker time.Ticker,
	reportTicker time.Ticker,
) {

	ch := make(chan map[string]any, 1000)
	defer close(ch)

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			logger.Log.Info("polling metrics...")
			produceGaugeMetrics(ch)
			produceCounterMetrics(ch)
		case <-reportTicker.C:
			logger.Log.Info("reporting metrics...")
			consumeMetrics(ctx, facade, ch)
		}
	}
}

func consumeMetrics(
	ctx context.Context,
	handler MetricFacade,
	ch chan map[string]any,
) {
	for {
		select {
		case m := <-ch:
			err := handler.Update(ctx, m)
			if err != nil {
				logger.Log.Error(err)
			}
		default:
			return
		}
	}
}

func produceGaugeMetrics(ch chan map[string]any) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	ch <- map[string]any{"type": types.GaugeMetricType, "name": "Alloc", "value": float64(memStats.Alloc)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "BuckHashSys", "value": float64(memStats.BuckHashSys)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "Frees", "value": float64(memStats.Frees)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "GCCPUFraction", "value": float64(memStats.GCCPUFraction)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "GCSys", "value": float64(memStats.GCSys)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "HeapAlloc", "value": float64(memStats.HeapAlloc)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "HeapIdle", "value": float64(memStats.HeapIdle)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "HeapInuse", "value": float64(memStats.HeapInuse)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "HeapObjects", "value": float64(memStats.HeapObjects)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "HeapReleased", "value": float64(memStats.HeapReleased)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "HeapSys", "value": float64(memStats.HeapSys)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "LastGC", "value": float64(memStats.LastGC)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "Lookups", "value": float64(memStats.Lookups)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "MCacheInuse", "value": float64(memStats.MCacheInuse)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "MCacheSys", "value": float64(memStats.MCacheSys)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "MSpanInuse", "value": float64(memStats.MSpanInuse)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "MSpanSys", "value": float64(memStats.MSpanSys)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "Mallocs", "value": float64(memStats.Mallocs)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "NextGC", "value": float64(memStats.NextGC)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "NumForcedGC", "value": float64(memStats.NumForcedGC)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "NumGC", "value": float64(memStats.NumGC)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "OtherSys", "value": float64(memStats.OtherSys)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "PauseTotalNs", "value": float64(memStats.PauseTotalNs)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "StackInuse", "value": float64(memStats.StackInuse)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "StackSys", "value": float64(memStats.StackSys)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "Sys", "value": float64(memStats.Sys)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "TotalAlloc", "value": float64(memStats.TotalAlloc)}
	ch <- map[string]any{"type": types.GaugeMetricType, "name": "RandomValue", "value": float64(rand.Float64())}
}

func produceCounterMetrics(ch chan map[string]any) {
	ch <- map[string]any{"type": types.CounterMetricType, "name": "PollCount", "value": int64(1)}
}
