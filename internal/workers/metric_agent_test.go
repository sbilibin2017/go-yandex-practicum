package workers

import (
	"context"
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
	ch := make(chan types.MetricUpdatePathRequest, 100)

	// Ожидаем, что будут вызовы на отправку обновлений метрик
	mockFacade.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, metric types.MetricUpdatePathRequest) error {
		assert.NotEmpty(t, metric.Name)
		assert.NotEmpty(t, metric.Value)
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
	ch := make(chan types.MetricUpdatePathRequest, 1)

	mockFacade.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, metric types.MetricUpdatePathRequest) error {
		assert.Equal(t, "TestMetric", metric.Name)
		assert.Equal(t, "123", metric.Value)
		return nil
	}).Times(1)

	ch <- types.MetricUpdatePathRequest{
		Name:  "TestMetric",
		Value: "123",
		Type:  string(types.GaugeMetricType),
	}

	consumeMetrics(context.Background(), mockFacade, ch)
}

func TestProduceGaugeMetrics(t *testing.T) {
	t.Helper()

	ch := make(chan types.MetricUpdatePathRequest, 50)

	go produceGaugeMetrics(ch)

	time.Sleep(100 * time.Millisecond)

	assert.True(t, len(ch) > 0, "expected some gauge metrics to be produced")

	for i := 0; i < len(ch); i++ {
		metric := <-ch
		assert.NotEmpty(t, metric.Name)
		assert.NotEmpty(t, metric.Value)
		assert.Equal(t, string(types.GaugeMetricType), metric.Type)
	}
}

func TestProduceCounterMetrics(t *testing.T) {
	t.Helper()

	ch := make(chan types.MetricUpdatePathRequest, 1)

	go produceCounterMetrics(ch)

	select {
	case metric := <-ch:
		assert.Equal(t, "PollCount", metric.Name)
		assert.Equal(t, "1", metric.Value)
		assert.Equal(t, string(types.CounterMetricType), metric.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("No counter metric received")
	}
}
