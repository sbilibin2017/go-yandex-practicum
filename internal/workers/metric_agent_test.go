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

func TestConsumeMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ctx := context.Background()
	ch := make(chan map[string]any, 2)
	metric1 := map[string]any{"type": "gauge", "name": "Test1", "value": 123}
	metric2 := map[string]any{"type": "counter", "name": "Test2", "value": 456}
	ch <- metric1
	ch <- metric2
	mockFacade.EXPECT().Update(ctx, metric1).Return(nil)
	mockFacade.EXPECT().Update(ctx, metric2).Return(nil)
	consumeMetrics(ctx, mockFacade, ch)
}

func TestConsumeMetrics_ErrorHandled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ctx := context.Background()
	ch := make(chan map[string]any, 1)
	metric := map[string]any{"type": "gauge", "name": "TestError", "value": 999}
	ch <- metric
	expectedErr := errors.New("update failed")
	mockFacade.
		EXPECT().
		Update(ctx, metric).
		Return(expectedErr)

	consumeMetrics(ctx, mockFacade, ch)
}

func TestProduceGaugeMetrics(t *testing.T) {
	ch := make(chan map[string]any, 100)
	produceGaugeMetrics(ch)
	count := len(ch)
	assert.Greater(t, count, 0, "Gauge metrics should be produced")
	for i := 0; i < count; i++ {
		m := <-ch
		assert.Equal(t, types.GaugeMetricType, m["type"])
		assert.NotEmpty(t, m["name"])
	}
}

func TestProduceCounterMetrics(t *testing.T) {
	ch := make(chan map[string]any, 1)
	produceCounterMetrics(ch)
	m := <-ch
	assert.Equal(t, types.CounterMetricType, m["type"])
	assert.Equal(t, "PollCount", m["name"])
	assert.Equal(t, int64(1), m["value"])
}

func TestStartMetricAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ctx, cancel := context.WithCancel(context.Background())
	pollTicker := time.NewTicker(10 * time.Millisecond)
	reportTicker := time.NewTicker(20 * time.Millisecond)
	defer pollTicker.Stop()
	defer reportTicker.Stop()
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	mockFacade.EXPECT().Update(gomock.Any(), gomock.Any()).AnyTimes()
	StartMetricAgent(ctx, mockFacade, *pollTicker, *reportTicker)
}
