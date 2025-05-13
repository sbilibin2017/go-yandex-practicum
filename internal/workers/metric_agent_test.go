package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestStartMetricAgent(t *testing.T) {
	t.Helper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ch := make(chan types.Metrics, 100)
	mockFacade.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, metric types.Metrics) error {
		assert.NotEmpty(t, metric.ID)
		assert.NotEmpty(t, metric.Type)
		return nil
	}).AnyTimes()
	pollTicker := time.NewTicker(10 * time.Millisecond)
	reportTicker := time.NewTicker(15 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	agent := NewMetricAgent(mockFacade, ch, pollTicker, reportTicker)
	go agent.Start(ctx)
	time.Sleep(100 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestConsumeMetrics(t *testing.T) {
	t.Helper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ch := make(chan types.Metrics, 1)
	mockFacade.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, metric types.Metrics) error {
		assert.Equal(t, "TestMetric", metric.ID)
		assert.Equal(t, string(types.GaugeMetricType), string(metric.Type))
		assert.NotNil(t, metric.Value)
		assert.Equal(t, 123.0, *metric.Value)
		return nil
	}).Times(1)
	ch <- types.Metrics{
		MetricID: types.MetricID{
			ID:   "TestMetric",
			Type: types.GaugeMetricType,
		},
		Value: func() *float64 { v := 123.0; return &v }(),
	}
	consumeMetrics(context.Background(), mockFacade, ch)
}

func TestProduceGaugeMetrics(t *testing.T) {
	t.Helper()
	ch := make(chan types.Metrics, 50)
	go produceGaugeMetrics(ch)
	time.Sleep(100 * time.Millisecond)
	assert.True(t, len(ch) > 0, "expected some gauge metrics to be produced")
	for i := 0; i < len(ch); i++ {
		metric := <-ch
		assert.NotEmpty(t, metric.ID)
		assert.Equal(t, string(types.GaugeMetricType), string(metric.Type))
		assert.NotNil(t, metric.Value)
	}
}

func TestProduceCounterMetrics(t *testing.T) {
	t.Helper()
	ch := make(chan types.Metrics, 1)
	go produceCounterMetrics(ch)
	select {
	case metric := <-ch:
		assert.Equal(t, "PollCount", metric.ID)
		assert.Equal(t, string(types.CounterMetricType), string(metric.Type))
		assert.NotNil(t, metric.Delta)
		assert.Equal(t, int64(1), *metric.Delta)
	case <-time.After(1 * time.Second):
		t.Fatal("No counter metric received")
	}
}

func TestConsumeMetrics_ErrorUpdatingMetric(t *testing.T) {
	t.Helper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ch := make(chan types.Metrics, 1)
	mockFacade.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, metric types.Metrics) error {
		assert.Equal(t, "TestMetric", metric.ID)
		assert.Equal(t, string(types.GaugeMetricType), string(metric.Type))
		assert.NotNil(t, metric.Value)
		assert.Equal(t, 123.0, *metric.Value)
		return errors.New("update failed")
	}).Times(1)
	ch <- types.Metrics{
		MetricID: types.MetricID{
			ID:   "TestMetric",
			Type: types.GaugeMetricType,
		},
		Value: func() *float64 { v := 123.0; return &v }(),
	}
	consumeMetrics(context.Background(), mockFacade, ch)
}
