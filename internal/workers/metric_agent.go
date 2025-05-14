package workers

import (
	"context"
	"math/rand"
	"runtime"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"go.uber.org/zap"
)

type MetricFacade interface {
	Update(ctx context.Context, metric types.Metrics) error
}

func StartMetricAgentWorker(
	ctx context.Context,
	metricFacade MetricFacade,
	metricCh chan types.Metrics,
	pollTicker *time.Ticker,
	reportTicker *time.Ticker,
) error {
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Context done, stopping metric agent.")
			return nil
		case <-pollTicker.C:
			logger.Log.Info("Polling metrics...")
			produceGaugeMetrics(metricCh)
			produceCounterMetrics(metricCh)
		case <-reportTicker.C:
			logger.Log.Info("Reporting metrics...")
			consumeMetrics(ctx, metricFacade, metricCh)
		}
	}
}

func consumeMetrics(
	ctx context.Context,
	handler MetricFacade,
	ch chan types.Metrics,
) {
	for {
		select {
		case m := <-ch:
			err := handler.Update(ctx, m)
			if err != nil {
				logger.Log.Error("Error updating metric", zap.String("id", m.ID), zap.Error(err))
			}
		default:
			return
		}
	}
}

func produceGaugeMetrics(ch chan types.Metrics) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := map[string]float64{
		"Alloc":         float64(memStats.Alloc),
		"BuckHashSys":   float64(memStats.BuckHashSys),
		"Frees":         float64(memStats.Frees),
		"GCCPUFraction": memStats.GCCPUFraction,
		"GCSys":         float64(memStats.GCSys),
		"HeapAlloc":     float64(memStats.HeapAlloc),
		"HeapIdle":      float64(memStats.HeapIdle),
		"HeapInuse":     float64(memStats.HeapInuse),
		"HeapObjects":   float64(memStats.HeapObjects),
		"HeapReleased":  float64(memStats.HeapReleased),
		"HeapSys":       float64(memStats.HeapSys),
		"LastGC":        float64(memStats.LastGC),
		"Lookups":       float64(memStats.Lookups),
		"MCacheInuse":   float64(memStats.MCacheInuse),
		"MCacheSys":     float64(memStats.MCacheSys),
		"MSpanInuse":    float64(memStats.MSpanInuse),
		"MSpanSys":      float64(memStats.MSpanSys),
		"Mallocs":       float64(memStats.Mallocs),
		"NextGC":        float64(memStats.NextGC),
		"NumForcedGC":   float64(memStats.NumForcedGC),
		"NumGC":         float64(memStats.NumGC),
		"OtherSys":      float64(memStats.OtherSys),
		"PauseTotalNs":  float64(memStats.PauseTotalNs),
		"StackInuse":    float64(memStats.StackInuse),
		"StackSys":      float64(memStats.StackSys),
		"Sys":           float64(memStats.Sys),
		"TotalAlloc":    float64(memStats.TotalAlloc),
		"RandomValue":   rand.Float64(),
	}

	for name, val := range metrics {
		metric := types.Metrics{
			MetricID: types.MetricID{
				ID:   name,
				Type: types.GaugeMetricType,
			},
			Delta: nil,
			Value: &val,
		}
		ch <- metric
	}
}

func produceCounterMetrics(ch chan types.Metrics) {
	counterData := map[string]int64{
		"PollCount": 1,
	}

	for name, delta := range counterData {
		metric := types.Metrics{
			MetricID: types.MetricID{
				ID:   name,
				Type: types.CounterMetricType,
			},
			Delta: &delta,
			Value: nil,
		}
		ch <- metric
	}
}
