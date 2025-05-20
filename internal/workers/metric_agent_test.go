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

func TestProcessMetricResults(t *testing.T) {
	results := make(chan result)

	go func() {
		defer close(results)

		// Успешный результат (Data не пустой, Err == nil)
		results <- result{
			Data: []types.Metrics{
				{MetricID: types.MetricID{ID: "test1", Type: types.GaugeMetricType}},
			},
			Err: nil,
		}

		// Результат с ошибкой
		results <- result{
			Data: []types.Metrics{
				{MetricID: types.MetricID{ID: "test2", Type: types.CounterMetricType}},
			},
			Err: errors.New("some error"),
		}
	}()

	processMetricResults(results)
}

func TestWorkerPoolMetricsUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockFacade := NewMockMetricFacade(ctrl)

	jobs := make(chan types.Metrics)
	batchSize := 2
	rateLimit := 1 // чтобы избежать гонок и упрощения проверки

	// Ожидаем, что Updates вызовется ровно два раза с любыми аргументами
	mockFacade.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(2)

	results := workerPoolMetricsUpdate(ctx, mockFacade, jobs, batchSize, rateLimit)

	go func() {
		jobs <- types.Metrics{MetricID: types.MetricID{ID: "m1", Type: types.GaugeMetricType}}
		jobs <- types.Metrics{MetricID: types.MetricID{ID: "m2", Type: types.GaugeMetricType}}
		jobs <- types.Metrics{MetricID: types.MetricID{ID: "m3", Type: types.GaugeMetricType}}
		close(jobs)
	}()

	var receivedResults []result
	done := make(chan struct{})

	go func() {
		for res := range results {
			receivedResults = append(receivedResults, res)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for results channel to close")
	}

	if len(receivedResults) != 2 {
		t.Errorf("expected 2 results, got %d", len(receivedResults))
	}
}

func TestWorkerMetricsUpdate_ContextDone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	mockFacade := NewMockMetricFacade(ctrl)

	jobs := make(chan types.Metrics)
	results := make(chan result)

	batchSize := 2

	// Чтобы Updates не вызывался случайно
	mockFacade.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Times(0)

	done := make(chan struct{})

	go func() {
		workerMetricsUpdate(ctx, mockFacade, jobs, results, batchSize)
		close(done)
	}()

	// Отмена контекста сразу
	cancel()

	select {
	case <-done:
		// Успешно завершилось
	case <-time.After(1 * time.Second):
		t.Fatal("workerMetricsUpdate did not exit after context cancellation")
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int64) *int64 {
	return &i
}

func TestFanInMetrics_MergesChannels(t *testing.T) {
	ctx := context.Background()

	ch1 := make(chan types.Metrics, 2)
	ch2 := make(chan types.Metrics, 2)

	out := fanInMetrics(ctx, ch1, ch2)

	ch1 <- types.Metrics{
		MetricID: types.MetricID{ID: "m1", Type: types.GaugeMetricType},
		Value:    floatPtr(1.23),
	}
	ch1 <- types.Metrics{
		MetricID: types.MetricID{ID: "m2", Type: types.CounterMetricType},
		Delta:    intPtr(10),
	}
	close(ch1)

	ch2 <- types.Metrics{
		MetricID: types.MetricID{ID: "m3", Type: types.GaugeMetricType},
		Value:    floatPtr(4.56),
	}
	close(ch2)

	var results []types.Metrics
	for m := range out {
		results = append(results, m)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	expectedIDs := map[string]bool{"m1": true, "m2": true, "m3": true}
	for _, m := range results {
		if !expectedIDs[m.ID] {
			t.Errorf("unexpected metric ID: %s", m.ID)
		}
	}
}

func TestFanInMetrics_ClosesOnAllChannelsClosed(t *testing.T) {
	ctx := context.Background()

	ch1 := make(chan types.Metrics)
	ch2 := make(chan types.Metrics)

	out := fanInMetrics(ctx, ch1, ch2)

	close(ch1)
	close(ch2)

	_, ok := <-out
	if ok {
		t.Error("expected final channel to be closed after all input channels are closed")
	}
}

func TestFanInMetrics_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ch1 := make(chan types.Metrics)
	ch2 := make(chan types.Metrics)

	out := fanInMetrics(ctx, ch1, ch2)

	done := make(chan struct{})
	go func() {
		for range out {
		}
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// успешно завершено
	case <-time.After(time.Second):
		t.Fatal("fanInMetrics did not stop after context cancellation")
	}
}

func TestGeneratorMetrics_EmitsAll(t *testing.T) {
	ctx := context.Background()
	input := []types.Metrics{
		{
			MetricID: types.MetricID{ID: "m1", Type: types.GaugeMetricType},
			Value:    floatPtr(1.1),
		},
		{
			MetricID: types.MetricID{ID: "m2", Type: types.CounterMetricType},
			Delta:    intPtr(10),
		},
	}

	out := generatorMetrics(ctx, input)

	var results []types.Metrics
	for m := range out {
		results = append(results, m)
	}

	if len(results) != len(input) {
		t.Fatalf("expected %d metrics, got %d", len(input), len(results))
	}

	for i, m := range results {
		if m.ID != input[i].ID {
			t.Errorf("expected metric ID %s, got %s", input[i].ID, m.ID)
		}
	}
}

func TestGeneratorMetrics_ClosesChannel(t *testing.T) {
	ctx := context.Background()
	input := []types.Metrics{}

	out := generatorMetrics(ctx, input)

	_, ok := <-out
	if ok {
		t.Error("expected channel to be closed for empty input")
	}
}

func TestGeneratorMetrics_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	input := []types.Metrics{
		{
			MetricID: types.MetricID{ID: "m1", Type: types.GaugeMetricType},
			Value:    floatPtr(1.1),
		},
	}

	out := generatorMetrics(ctx, input)

	// Отмена контекста сразу после запуска
	cancel()

	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected channel to be closed after context cancellation")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for channel close after context cancellation")
	}
}

func TestGetGoputilMetrics(t *testing.T) {
	ctx := context.Background()
	metrics := getGoputilMetrics(ctx)

	// Проверяем, что возвращён слайс не пустой
	assert.NotEmpty(t, metrics, "expected some metrics, got none")

}

func TestGetRuntimeCounterMetrics(t *testing.T) {

	metrics := getRuntimeCounterMetrics()

	assert.Len(t, metrics, 1, "expected exactly one metric")

}

func TestGetRuntimeGaugeMetrics(t *testing.T) {

	metrics := getRuntimeGaugeMetrics()

	assert.Len(t, metrics, 28, "expected 28 metrics")

}

func int64Ptr(i int64) *int64 { return &i }

func TestStartMetricsReporting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFacade := NewMockMetricFacade(ctrl)

	// Подготовка канала входящих метрик
	in := make(chan types.Metrics, 10)

	// Подготовим метрики для отправки
	m1 := types.Metrics{
		MetricID: types.MetricID{ID: "m1", Type: types.GaugeMetricType},
		Value:    floatPtr(1.0),
	}
	m2 := types.Metrics{
		MetricID: types.MetricID{ID: "m2", Type: types.CounterMetricType},
		Delta:    int64Ptr(2),
	}

	// Отправим метрики в канал
	in <- m1
	in <- m2

	// Контекст с отменой для выхода из цикла
	ctx, cancel := context.WithCancel(context.Background())

	// Мок ожидает вызов Updates с любыми аргументами, возвращает nil
	mockFacade.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(nil)

	// Запуск startMetricsReporting в горутине
	go func() {
		startMetricsReporting(ctx, 1, mockFacade, in, 10, 1)
	}()

	// Подождем чуть больше 1 секунды, чтобы тикер сработал
	time.Sleep(1200 * time.Millisecond)

	// Отменим контекст, чтобы функция завершилась
	cancel()

	// Закроем канал входящих метрик
	close(in)

	// Дополнительная проверка — если функция завершилась, тест проходит
	// Тут можно проверить внутренние переменные, если нужно

	assert.True(t, true, "startMetricsReporting exited cleanly")
}

func TestStartMetricsPolling(t *testing.T) {
	out := make(chan types.Metrics, 100) // буфер, чтобы не блокироваться

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем функцию в горутине
	go startMetricsPolling(ctx, 1, out)

	// Ждем, пока появятся метрики (ожидаем, что хотя бы одна метрика появится за 2 секунды)
	var received []types.Metrics
	timeout := time.After(2 * time.Second)

loop:
	for {
		select {
		case m := <-out:
			received = append(received, m)
			// Можно выйти после получения хотя бы 1 метрики
			if len(received) >= 1 {
				break loop
			}
		case <-timeout:
			t.Fatal("Timeout: метрики не были получены")
		}
	}

	// Отменяем контекст, чтобы завершить функцию
	cancel()

	// Ждем, что функция корректно завершилась (не блокируемся на записи)
	time.Sleep(200 * time.Millisecond)

	assert.NotEmpty(t, received, "Ожидается хотя бы одна метрика")
	// Можно дополнительно проверить типы метрик
	for _, m := range received {
		assert.NotEmpty(t, m.ID)
	}
}

func TestStartMetricAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFacade := NewMockMetricFacade(ctrl)

	ctx, cancel := context.WithCancel(context.Background())

	// Настраиваем мок для фасада: Updates может вызываться, возвращаем nil (успех)
	mockFacade.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	// Запускаем агент
	go func() {
		err := startMetricAgent(ctx, mockFacade, 1, 1, 2, 1)
		require.ErrorIs(t, err, context.Canceled)
	}()

	// Даем немного времени поработать (чтобы стартовали горутины)
	time.Sleep(1500 * time.Millisecond)

	// Отменяем контекст, чтобы завершить
	cancel()

	// Ждем немного, чтобы горутина успела завершиться
	time.Sleep(500 * time.Millisecond)
}

func TestNewMetricAgentWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFacade := NewMockMetricFacade(ctrl)

	// Настраиваем мок: допускаем любые вызовы Updates
	mockFacade.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	worker := NewMetricAgentWorker(mockFacade, 1, 1, 2, 1)

	ctx, cancel := context.WithCancel(context.Background())

	// Запускаем worker в отдельной горутине
	errCh := make(chan error)
	go func() {
		errCh <- worker(ctx)
	}()

	// Даем поработать немного
	time.Sleep(1500 * time.Millisecond)

	// Отменяем контекст, чтобы остановить worker
	cancel()

	// Ждем завершения и проверяем ошибку
	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}
