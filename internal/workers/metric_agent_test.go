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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ch := make(chan types.MetricUpdatePathRequest, 10)
	mockFacade.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, metric types.MetricUpdatePathRequest) error {
		assert.NotEmpty(t, metric.Name, "Metric Name should not be empty")
		assert.NotEmpty(t, metric.Value, "Metric Value should not be empty")
		return nil
	}).AnyTimes()
	pollTicker := time.NewTicker(10 * time.Millisecond)
	reportTicker := time.NewTicker(15 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go StartMetricAgent(ctx, mockFacade, ch, *pollTicker, *reportTicker)
	time.Sleep(100 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestConsumeMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ch := make(chan types.MetricUpdatePathRequest, 10)
	mockFacade.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, metric types.MetricUpdatePathRequest) error {
		assert.NotEmpty(t, metric.Name, "metric name should not be empty")
		assert.NotEmpty(t, metric.Value, "metric value should not be empty")
		return nil
	}).Times(1)
	ch <- types.MetricUpdatePathRequest{Name: "TestMetric", Value: "123"}
	consumeMetrics(context.Background(), mockFacade, ch)
}

func TestProduceGaugeMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ch := make(chan types.MetricUpdatePathRequest, 10)
	mockFacade.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, metric types.MetricUpdatePathRequest) error {
		assert.NotEmpty(t, metric.Name, "metric name should not be empty")
		assert.NotEmpty(t, metric.Value, "metric value should not be empty")
		return nil
	}).AnyTimes()
	go produceGaugeMetrics(ch)
	time.Sleep(500 * time.Millisecond)
	select {
	case metric := <-ch:
		assert.NotEmpty(t, metric.Name, "metric name should not be empty")
		assert.NotEmpty(t, metric.Value, "metric value should not be empty")
	case <-time.After(1 * time.Second):
		t.Fatal("no metrics received within timeout")
	}
}
