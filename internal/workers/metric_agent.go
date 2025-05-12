package workers

import (
	"context"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"go.uber.org/zap"
)

type MetricFacade interface {
	Update(ctx context.Context, metric types.MetricUpdatePathRequest) error
}

type MetricAgent struct {
	metricFacade MetricFacade
	metricCh     chan types.MetricUpdatePathRequest
	pollTicker   *time.Ticker
	reportTicker *time.Ticker
}

func NewMetricAgent(
	metricFacade MetricFacade,
	metricCh chan types.MetricUpdatePathRequest,
	pollTicker *time.Ticker,
	reportTicker *time.Ticker,
) *MetricAgent {
	return &MetricAgent{
		metricFacade: metricFacade,
		metricCh:     metricCh,
		pollTicker:   pollTicker,
		reportTicker: reportTicker,
	}
}

func (ma *MetricAgent) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Context done, stopping metric agent.")
			return
		case <-ma.pollTicker.C:
			logger.Log.Info("Polling metrics...")
			produceGaugeMetrics(ma.metricCh)
			produceCounterMetrics(ma.metricCh)
		case <-ma.reportTicker.C:
			logger.Log.Info("Reporting metrics...")
			consumeMetrics(ctx, ma.metricFacade, ma.metricCh)
		}
	}
}

func consumeMetrics(
	ctx context.Context,
	handler MetricFacade,
	ch chan types.MetricUpdatePathRequest,
) {
	for {
		select {
		case m := <-ch:
			logger.Log.Info("Consuming metric", zap.String("name", m.Name), zap.String("type", m.Type), zap.String("value", m.Value))
			err := handler.Update(ctx, m)
			if err != nil {
				logger.Log.Error("Error updating metric", zap.String("name", m.Name), zap.Error(err))
			} else {
				logger.Log.Info("Successfully updated metric", zap.String("name", m.Name))
			}
		default:
			logger.Log.Debug("No metrics to consume.")
			return
		}
	}
}

func produceGaugeMetrics(ch chan types.MetricUpdatePathRequest) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	logger.Log.Info("Producing gauge metrics...")
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "Alloc", Value: strconv.FormatFloat(float64(memStats.Alloc), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "BuckHashSys", Value: strconv.FormatFloat(float64(memStats.BuckHashSys), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "Frees", Value: strconv.FormatFloat(float64(memStats.Frees), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "GCCPUFraction", Value: strconv.FormatFloat(memStats.GCCPUFraction, 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "GCSys", Value: strconv.FormatFloat(float64(memStats.GCSys), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "HeapAlloc", Value: strconv.FormatFloat(float64(memStats.HeapAlloc), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "HeapIdle", Value: strconv.FormatFloat(float64(memStats.HeapIdle), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "HeapInuse", Value: strconv.FormatFloat(float64(memStats.HeapInuse), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "HeapObjects", Value: strconv.FormatFloat(float64(memStats.HeapObjects), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "HeapReleased", Value: strconv.FormatFloat(float64(memStats.HeapReleased), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "HeapSys", Value: strconv.FormatFloat(float64(memStats.HeapSys), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "LastGC", Value: strconv.FormatFloat(float64(memStats.LastGC), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "Lookups", Value: strconv.FormatFloat(float64(memStats.Lookups), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "MCacheInuse", Value: strconv.FormatFloat(float64(memStats.MCacheInuse), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "MCacheSys", Value: strconv.FormatFloat(float64(memStats.MCacheSys), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "MSpanInuse", Value: strconv.FormatFloat(float64(memStats.MSpanInuse), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "MSpanSys", Value: strconv.FormatFloat(float64(memStats.MSpanSys), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "Mallocs", Value: strconv.FormatFloat(float64(memStats.Mallocs), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "NextGC", Value: strconv.FormatFloat(float64(memStats.NextGC), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "NumForcedGC", Value: strconv.FormatFloat(float64(memStats.NumForcedGC), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "NumGC", Value: strconv.FormatFloat(float64(memStats.NumGC), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "OtherSys", Value: strconv.FormatFloat(float64(memStats.OtherSys), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "PauseTotalNs", Value: strconv.FormatFloat(float64(memStats.PauseTotalNs), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "StackInuse", Value: strconv.FormatFloat(float64(memStats.StackInuse), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "StackSys", Value: strconv.FormatFloat(float64(memStats.StackSys), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "Sys", Value: strconv.FormatFloat(float64(memStats.Sys), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "TotalAlloc", Value: strconv.FormatFloat(float64(memStats.TotalAlloc), 'f', -1, 64)}
	ch <- types.MetricUpdatePathRequest{Type: string(types.GaugeMetricType), Name: "RandomValue", Value: strconv.FormatFloat(rand.Float64(), 'f', -1, 64)}
	logger.Log.Info("Gauge metrics produced.")
}

func produceCounterMetrics(ch chan types.MetricUpdatePathRequest) {
	logger.Log.Info("Producing counter metrics...")
	ch <- types.MetricUpdatePathRequest{Type: string(types.CounterMetricType), Name: "PollCount", Value: "1"}
	logger.Log.Info("Counter metrics produced.")
}
