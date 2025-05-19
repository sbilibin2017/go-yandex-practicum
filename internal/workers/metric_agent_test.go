package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricAgentWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFacade := NewMockMetricFacade(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ожидаем, что Updates может вызываться (или не вызываться) — в зависимости от реализации
	mockFacade.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	worker := NewMetricAgentWorker(mockFacade, 1, 2, 2, 5)

	errCh := make(chan error)
	go func() {
		errCh <- worker(ctx)
	}()

	time.Sleep(3 * time.Second)
	cancel() // отменяем контекст, чтобы завершить worker

	err := <-errCh
	if err != nil && err != context.Canceled {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStartMetricAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFacade := NewMockMetricFacade(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockFacade.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	errCh := make(chan error)
	go func() {
		err := startMetricAgent(ctx, mockFacade, 1, 2, 2, 5)
		errCh <- err
	}()

	time.Sleep(5 * time.Second)
	cancel()

	err := <-errCh
	// Проверяем, что ошибка либо nil, либо context.Canceled (т.е. отмена ожидаема)
	if err != nil && err != context.Canceled {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessMetricResults(t *testing.T) {
	t.Run("process success and error results", func(t *testing.T) {
		results := make(chan result, 2)

		metric := types.Metrics{
			MetricID: types.MetricID{ID: "test", Type: types.GaugeMetricType},
			Value:    nil,
		}

		// Положим один успешный результат
		results <- result{Data: []types.Metrics{metric}, Err: nil}
		// И один с ошибкой
		results <- result{Data: []types.Metrics{metric}, Err: errors.New("some error")}
		close(results)

		// Просто вызываем функцию — она должна корректно обработать канал и завершиться
		processMetricResults(results)

		// Проверок на выход нет, главное чтобы не было паники
		assert.True(t, true, "processMetricResults should complete without panic")
	})
}

func TestGeneratorMetrics_AllMetricsSent(t *testing.T) {
	ctx := context.Background()

	mockMetrics := []types.Metrics{
		{MetricID: types.MetricID{ID: "m1"}},
		{MetricID: types.MetricID{ID: "m2"}},
	}

	inputFunc := func(ctx context.Context) []types.Metrics {
		return mockMetrics
	}

	out := generatorMetrics(ctx, inputFunc)

	var result []string
	for m := range out {
		result = append(result, m.MetricID.ID)
	}

	assert.ElementsMatch(t, []string{"m1", "m2"}, result)
}

func TestGeneratorMetrics_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	inputFunc := func(ctx context.Context) []types.Metrics {
		// Имитация задержки — контекст отменится до возврата
		time.Sleep(50 * time.Millisecond)
		return []types.Metrics{
			{MetricID: types.MetricID{ID: "should_not_send"}},
		}
	}

	cancel() // отменяем до вызова generatorMetrics

	out := generatorMetrics(ctx, inputFunc)

	var received []types.Metrics
	for m := range out {
		received = append(received, m)
	}

	assert.Empty(t, received, "output channel should be empty after context cancel")
}

func TestGetGoputilMetrics(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	metrics := getGoputilMetrics(ctx)

	assert.NotEmpty(t, metrics, "Expected some metrics to be collected")

}

func TestGetRuntimeCounterMetrics(t *testing.T) {
	metrics := getRuntimeCounterMetrics(context.Background())

	require.Len(t, metrics, 1, "should return exactly one metric")

}

func TestGetRuntimeGaugeMetrics(t *testing.T) {
	metrics := getRuntimeGaugeMetrics(context.Background())

	// Ожидаем ровно 28 метрик
	require.Len(t, metrics, 28)
}

func TestWorkerMetricsUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFacade := NewMockMetricFacade(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobCh := make(chan types.Metrics, 10)
	resultsCh := make(chan result, 10)

	batchSize := 3

	// Ожидаем вызов Updates с батчем из 3 метрик, возвращаем nil
	mockFacade.EXPECT().Updates(gomock.Any(), gomock.Len(batchSize)).Return(nil).Times(1)

	// Запускаем воркер
	go workerMetricsUpdate(ctx, mockFacade, jobCh, resultsCh, batchSize)

	// Отправляем в job канал 3 метрики
	for i := 0; i < batchSize; i++ {
		m := types.Metrics{
			MetricID: types.MetricID{ID: "test_metric", Type: types.GaugeMetricType},
			Value:    new(float64),
		}
		jobCh <- m
	}
	// Закрываем канал, чтобы воркер вышел после обработки
	close(jobCh)

	// Ждем результат
	select {
	case res := <-resultsCh:
		assert.NoError(t, res.Err)
		assert.Len(t, res.Data, batchSize)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for worker result")
	}
}

// Простая заглушка воркера для теста, который просто читает из job и кладет результат без ошибки
func dummyWorker(
	ctx context.Context,
	facade MetricFacade,
	job chan types.Metrics,
	results chan result,
	batchSize int,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-job:
			if !ok {
				return
			}

			// В тесте просто возвращаем батч из одного элемента, без вызова фасада
			results <- result{
				Data: []types.Metrics{m},
				Err:  nil,
			}
		}
	}
}

func TestWorkerPoolMetricsUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFacade := NewMockMetricFacade(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan types.Metrics, 10)

	numWorkers := 3
	batchSize := 1

	// Запускаем workerPoolMetricsUpdate с dummyWorker
	resultsCh := workerPoolMetricsUpdate(ctx, mockFacade, dummyWorker, jobs, numWorkers, batchSize)

	// Отправим несколько задач
	testMetrics := types.Metrics{
		MetricID: types.MetricID{ID: "test_metric", Type: types.GaugeMetricType},
	}
	for i := 0; i < 5; i++ {
		jobs <- testMetrics
	}
	close(jobs)

	var results []result
	timeout := time.After(2 * time.Second)

loop:
	for {
		select {
		case res, ok := <-resultsCh:
			if !ok {
				break loop
			}
			results = append(results, res)
		case <-timeout:
			t.Fatal("timeout waiting for results")
		}
	}

	// Проверяем, что получили 5 результатов
	assert.Len(t, results, 5)

	// Проверяем, что ошибок нет
	for _, r := range results {
		assert.NoError(t, r.Err)
	}
}

func TestFanInMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch1 := make(chan types.Metrics, 2)
	ch2 := make(chan types.Metrics, 2)

	// Заполняем каналы тестовыми метриками
	m1 := types.Metrics{MetricID: types.MetricID{ID: "m1", Type: types.GaugeMetricType}}
	m2 := types.Metrics{MetricID: types.MetricID{ID: "m2", Type: types.GaugeMetricType}}
	m3 := types.Metrics{MetricID: types.MetricID{ID: "m3", Type: types.GaugeMetricType}}

	ch1 <- m1
	ch1 <- m2
	close(ch1)

	ch2 <- m3
	close(ch2)

	// Вызываем функцию fanInMetrics
	outCh := fanInMetrics(ctx, ch1, ch2)

	var got []types.Metrics

	timeout := time.After(1 * time.Second)
loop:
	for {
		select {
		case metric, ok := <-outCh:
			if !ok {
				break loop
			}
			got = append(got, metric)
		case <-timeout:
			t.Fatal("timeout waiting for fanInMetrics output")
		}
	}

	// Проверяем, что все метрики из обоих каналов пришли во внешний канал
	assert.ElementsMatch(t, got, []types.Metrics{m1, m2, m3})
}
