package workers

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRuntimeGaugeMetrics(t *testing.T) {
	metrics := generateRuntimeGaugeMetrics(context.Background())
	assert.NotEmpty(t, metrics)
	var allocMetric *types.Metrics
	for _, m := range metrics {
		if m.MetricID.ID == "Alloc" {
			allocMetric = &m
			break
		}
	}
	assert.NotNil(t, allocMetric)
	assert.Equal(t, types.GaugeMetricType, allocMetric.MetricID.Type)
	assert.NotNil(t, allocMetric.Value)
	assert.GreaterOrEqual(t, *allocMetric.Value, 0.0)
}

func TestGenerateRuntimeCounterMetrics(t *testing.T) {
	metrics := generateRuntimeCounterMetrics(context.Background())
	assert.NotEmpty(t, metrics)
	var pollCountMetric *types.Metrics
	for _, m := range metrics {
		if m.MetricID.ID == "PollCount" {
			pollCountMetric = &m
			break
		}
	}
	assert.NotNil(t, pollCountMetric)
	assert.Equal(t, types.CounterMetricType, pollCountMetric.MetricID.Type)
	assert.NotNil(t, pollCountMetric.Delta)
	assert.Equal(t, int64(1), *pollCountMetric.Delta)
}

func TestGenerateGopsutilGaugeMetrics(t *testing.T) {
	ctx := context.Background()
	metrics := generateGopsutilGaugeMetrics(ctx)
	assert.NotEmpty(t, metrics)
	var totalMem, freeMem *types.Metrics
	for _, m := range metrics {
		if m.MetricID.ID == "TotalMemory" {
			totalMem = &m
		}
		if m.MetricID.ID == "FreeMemory" {
			freeMem = &m
		}
	}
	assert.NotNil(t, totalMem)
	assert.NotNil(t, totalMem.Value)
	assert.GreaterOrEqual(t, *totalMem.Value, 0.0)
	assert.NotNil(t, freeMem)
	assert.NotNil(t, freeMem.Value)
	assert.GreaterOrEqual(t, *freeMem.Value, 0.0)
	cpuFound := false
	for _, m := range metrics {
		if len(m.MetricID.ID) > len("CPUutilization") && m.MetricID.ID[:len("CPUutilization")] == "CPUutilization" {
			cpuFound = true
			assert.NotNil(t, m.Value)
			assert.GreaterOrEqual(t, *m.Value, 0.0)
			_, err := strconv.Atoi(m.MetricID.ID[len("CPUutilization"):])
			assert.NoError(t, err)
		}
	}
	assert.True(t, cpuFound)
}

func TestGeneratorGopsutilMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pollInterval := 1
	metricsChan := generatorGopsutilMetrics(ctx, pollInterval)
	var received []types.Metrics
	timeout := time.After(5 * time.Second)
loop:
	for {
		select {
		case m, ok := <-metricsChan:
			if !ok {
				break loop
			}
			received = append(received, m)
			if len(received) >= 3 {
				break loop
			}
		case <-timeout:
			t.Fatal("Timeout waiting for metrics")
		}
	}
	assert.NotEmpty(t, received)
	cancel()
	timeoutClose := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-metricsChan:
			if !ok {
				return
			}
		case <-timeoutClose:
			t.Fatal("Timeout waiting for channel to close after cancel")
		}
	}
}

func TestGeneratorRuntimeMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pollInterval := 1
	metricsChan := generatorRuntimeMetrics(ctx, pollInterval)
	var received []types.Metrics
	timeout := time.After(5 * time.Second)
loop:
	for {
		select {
		case m, ok := <-metricsChan:
			if !ok {
				break loop
			}
			received = append(received, m)
			if len(received) >= 5 {
				break loop
			}
		case <-timeout:
			t.Fatal("Timeout waiting for metrics")
		}
	}
	assert.NotEmpty(t, received)
	cancel()
	timeoutClose := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-metricsChan:
			if !ok {
				return
			}
		case <-timeoutClose:
			t.Fatal("Timeout waiting for channel to close after cancel")
		}
	}
}

func TestFanIn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch1 := make(chan types.Metrics)
	ch2 := make(chan types.Metrics)
	out := fanIn(ctx, ch1, ch2)
	go func() {
		ch1 <- types.Metrics{MetricID: types.MetricID{ID: "m1"}, Value: ptrFloat64(1)}
		ch1 <- types.Metrics{MetricID: types.MetricID{ID: "m2"}, Value: ptrFloat64(2)}
		close(ch1)
	}()
	go func() {
		ch2 <- types.Metrics{MetricID: types.MetricID{ID: "m3"}, Value: ptrFloat64(3)}
		ch2 <- types.Metrics{MetricID: types.MetricID{ID: "m4"}, Value: ptrFloat64(4)}
		close(ch2)
	}()
	var received []types.Metrics
	for m := range out {
		received = append(received, m)
	}
	assert.Len(t, received, 4)
	_, ok := <-out
	assert.False(t, ok)
}

func ptrFloat64(f float64) *float64 {
	return &f
}

func TestFanOutWorker_BatchSend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFacade := NewMockMetricFacade(ctrl)
	mockSemaphore := NewMockSemaphore(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inputCh := make(chan types.Metrics)
	resultCh := make(chan Result, 10)

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{ID: "metric1", Type: types.GaugeMetricType},
			Value:    floatPtr(1.1),
		},
		{
			MetricID: types.MetricID{ID: "metric2", Type: types.GaugeMetricType},
			Value:    floatPtr(2.2),
		},
	}

	mockSemaphore.EXPECT().Acquire(gomock.Any(), int64(1)).Return(nil)
	mockSemaphore.EXPECT().Release(int64(1))
	mockFacade.EXPECT().Updates(gomock.Any(), gomock.Len(2)).Return(errors.New("update error"))

	go fanOutWorker(ctx, mockFacade, mockSemaphore, inputCh, 2, 1, resultCh)

	// Отправляем метрики
	inputCh <- metrics[0]
	inputCh <- metrics[1]

	// Ждём немного, чтобы воркер обработал батч
	time.Sleep(100 * time.Millisecond)
	close(inputCh)

	// Проверяем результат ошибок
	var gotResults []Result
	for i := 0; i < 2; i++ {
		select {
		case res := <-resultCh:
			gotResults = append(gotResults, res)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for results")
		}
	}

	require.Len(t, gotResults, 2)
	for _, res := range gotResults {
		require.EqualError(t, res.err, "update error")
	}
}

func TestFanOutWorkerPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFacade := NewMockMetricFacade(ctrl)
	mockSemaphore := NewMockSemaphore(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inputCh := make(chan types.Metrics)
	resultCh := make(chan Result, 10)

	batchSize := 2
	reportInterval := 1
	workerCount := 1

	var wg sync.WaitGroup
	wg.Add(1)

	// Ожидаем вызовы Acquire -> Updates -> Release в порядке
	gomock.InOrder(
		mockSemaphore.EXPECT().Acquire(gomock.Any(), int64(1)).Return(nil),
		mockFacade.EXPECT().Updates(gomock.Any(), gomock.Len(batchSize)).Do(func(_ context.Context, _ []types.Metrics) {
			// Как только Updates вызывается — снимаем ожидание
			wg.Done()
		}).Return(nil),
		mockSemaphore.EXPECT().Release(int64(1)),
	)

	go fanOutWorkerPool(ctx, mockFacade, mockSemaphore, inputCh, workerCount, batchSize, reportInterval, resultCh)

	// Отправляем два метрики, чтобы сработал batch
	inputCh <- types.Metrics{MetricID: types.MetricID{ID: "m1", Type: types.GaugeMetricType}}
	inputCh <- types.Metrics{MetricID: types.MetricID{ID: "m2", Type: types.CounterMetricType}}

	// Ждем вызова Updates (то есть успешной обработки batch)
	wg.Wait()

	// Закрываем inputCh, чтобы воркер завершился
	close(inputCh)

	// Читаем все из resultCh, чтобы горутины не блокировались
	for range resultCh {
	}

	// Отмена контекста, если надо
	cancel()
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestStartMetricAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFacade := NewMockMetricFacade(ctrl)
	mockSemaphore := NewMockSemaphore(ctrl)

	ctx, cancel := context.WithCancel(context.Background())

	// Разрешаем любые вызовы Acquire и Release, чтобы не блокировать поток
	mockSemaphore.EXPECT().Acquire(gomock.Any(), int64(1)).Return(nil).AnyTimes()
	mockSemaphore.EXPECT().Release(int64(1)).AnyTimes()

	// Разрешаем любые вызовы Updates, возвращаем nil, AnyTimes, чтобы не мешать тесту
	mockFacade.EXPECT().Updates(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	errCh := make(chan error, 1)

	go func() {
		err := startMetricAgent(ctx, mockFacade, mockSemaphore, 1, 2, 2, 1)
		errCh <- err
	}()

	// Даём немного времени на запуск и работу
	time.Sleep(100 * time.Millisecond)

	// Отменяем контекст - должны корректно завершить работу
	cancel()

	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}

func TestNewMetricAgentWorkerAndStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFacade := NewMockMetricFacade(ctrl)
	mockSemaphore := NewMockSemaphore(ctrl)

	// Разрешаем любые вызовы Acquire и Release для корректной работы воркера
	mockSemaphore.EXPECT().Acquire(gomock.Any(), int64(1)).Return(nil).AnyTimes()
	mockSemaphore.EXPECT().Release(int64(1)).AnyTimes()

	// Разрешаем любые вызовы Updates, возвращаем nil
	mockFacade.EXPECT().Updates(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	worker := NewMetricAgentWorker(
		mockFacade,
		mockSemaphore,
		1, // pollInterval
		2, // workerCount
		2, // batchSize
		1, // reportInterval
	)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		err := worker.Start(ctx)
		errCh <- err
	}()

	time.Sleep(100 * time.Millisecond) // Дать немного времени на работу

	cancel() // Отменяем контекст, чтобы остановить работу

	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}
