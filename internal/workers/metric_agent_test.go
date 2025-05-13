package workers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestConsumeMetricsError(t *testing.T) {
	t.Helper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок для MetricFacade
	mockFacade := NewMockMetricFacade(ctrl)
	ch := make(chan types.Metrics, 1)

	// Настроим мок на возврат ошибки
	mockFacade.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, metric types.Metrics) error {
		assert.NotEmpty(t, metric.ID)
		assert.NotEmpty(t, metric.Type)
		return fmt.Errorf("mock error")
	}).Times(1)

	// Отправляем метрику в канал
	ch <- types.Metrics{
		MetricID: types.MetricID{
			ID:   "PollCount",
			Type: types.CounterMetricType,
		},
		Delta: func() *int64 { v := int64(1); return &v }(),
	}

	// Создаем канал для завершения горутины
	done := make(chan bool)

	// Запускаем consumeMetrics в горутине
	go func() {
		consumeMetrics(context.Background(), mockFacade, ch)
		done <- true
	}()

	// Ожидаем завершения работы горутины
	<-done
	// Проверяем, что канал пуст
	assert.True(t, len(ch) == 0, "Expected the metric to be consumed")
}

func TestPollTickerCase(t *testing.T) {
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

	// Установим интервалы
	pollInterval := 1
	reportInterval := 2
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для завершения
	done := make(chan bool)

	// Запускаем StartMetricAgentWorker в горутине
	go func() {
		defer close(done)
		err := StartMetricAgentWorker(ctx, mockFacade, ch, time.NewTicker(time.Duration(pollInterval)*time.Second), time.NewTicker(time.Duration(reportInterval)*time.Second))
		assert.NoError(t, err)
	}()

	// Ждем, чтобы горутина могла выполнить операцию
	time.Sleep(2 * time.Second)

	// Проверяем, что метрики были произведены
	assert.True(t, len(ch) > 0, "Expected metrics to be produced by pollTicker")

	// Отменяем контекст
	cancel()
	<-done // Ждем завершения работы горутины
}

func TestReportTickerCase(t *testing.T) {
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

	// Установим интервалы
	pollInterval := 1
	reportInterval := 2
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем StartMetricAgentWorker
	go StartMetricAgentWorker(ctx, mockFacade, ch, time.NewTicker(time.Duration(pollInterval)*time.Second), time.NewTicker(time.Duration(reportInterval)*time.Second))

	// Даем время на выполнение
	time.Sleep(2 * time.Second)

	// Проверяем, что метрики были произведены
	assert.True(t, len(ch) > 0, "Expected metrics to be produced and consumed by reportTicker")

	// Отменяем контекст
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestPollTickerProduceMetrics(t *testing.T) {
	t.Helper()

	// Канал для метрик
	ch := make(chan types.Metrics, 100)

	// Запускаем производство метрик
	go produceGaugeMetrics(ch)
	go produceCounterMetrics(ch)

	// Даем время на выполнение
	time.Sleep(50 * time.Millisecond)

	// Проверяем, что метрики были произведены
	assert.True(t, len(ch) > 0, "Expected metrics to be produced by polling")
}

func TestReportTickerConsumeMetrics(t *testing.T) {
	t.Helper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFacade := NewMockMetricFacade(ctrl)
	ch := make(chan types.Metrics, 1)

	mockFacade.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, metric types.Metrics) error {
		assert.Equal(t, "PollCount", metric.ID)
		assert.Equal(t, string(types.CounterMetricType), string(metric.Type))
		assert.NotNil(t, metric.Delta)
		assert.Equal(t, int64(1), *metric.Delta)
		return nil
	}).Times(1)

	// Отправляем метрику в канал
	ch <- types.Metrics{
		MetricID: types.MetricID{
			ID:   "PollCount",
			Type: types.CounterMetricType,
		},
		Delta: func() *int64 { v := int64(1); return &v }(),
	}

	// Запускаем consumeMetrics
	go consumeMetrics(context.Background(), mockFacade, ch)

	// Даем время на выполнение
	time.Sleep(50 * time.Millisecond)

	// Проверяем, что канал пуст
	assert.True(t, len(ch) == 0, "Expected the metric to be consumed")
}
