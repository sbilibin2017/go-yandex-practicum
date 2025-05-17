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

func TestMetricsFanOut_SemaphoreAcquireError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockFacade := NewMockMetricFacade(ctrl)
	mockSema := NewMockSemaphore(ctrl)
	numWorkers := 3
	batchSize := 10
	inputCh := make(chan types.Metrics)
	mockSema.EXPECT().Acquire(gomock.Any(), int64(1)).Return(errors.New("acquire error")).Times(numWorkers)
	channels := metricsFanOut(ctx, mockFacade, mockSema, numWorkers, batchSize, inputCh)
	for i, ch := range channels {
		if ch != nil {
			t.Errorf("expected channel at index %d to be nil, but got %v", i, ch)
		}
	}
}

func TestStartMetricAgent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	facade := NewMockMetricFacade(ctrl)
	sema := NewMockSemaphore(ctrl)

	facade.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(nil)

	sema.EXPECT().
		Acquire(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(nil)

	sema.EXPECT().
		Release(gomock.Any()).
		AnyTimes()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		errCh <- NewMetricAgentWorker(
			ctx,
			facade,
			sema,
			1, // poll interval 1 second
			1, // report interval 1 second
			1,
			1,
		)(ctx)
	}()

	time.Sleep(1500 * time.Millisecond) // Ждем немного, чтобы worker поработал

	cancel() // Отменяем контекст

	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}

func TestWaitForErrors(t *testing.T) {
	t.Run("error received", func(t *testing.T) {
		errCh := make(chan error, 1)
		errCh <- errors.New("some error")
		close(errCh)

		ctx := context.Background()
		err := waitForErrors(ctx, errCh)
		require.Error(t, err)
		require.EqualError(t, err, "some error")
	})
	t.Run("channel closed without error", func(t *testing.T) {
		errCh := make(chan error)
		close(errCh)

		ctx := context.Background()
		err := waitForErrors(ctx, errCh)
		require.NoError(t, err)
	})
	t.Run("context cancelled", func(t *testing.T) {
		errCh := make(chan error)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := waitForErrors(ctx, errCh)
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestGeneratorRuntimeGaugeMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := generatorRuntimeGaugeMetrics(ctx)
	for i := 0; i < 5; i++ {
		select {
		case m, ok := <-ch:
			assert.True(t, ok, "канал должен быть открыт")
			assert.NotNil(t, m.Value, "метрика должна иметь значение")
		case <-time.After(time.Second):
			t.Fatal("таймаут ожидания метрики")
		}
	}
	cancel()
	timeout := time.After(time.Second)
	closed := false
	for !closed {
		select {
		case _, ok := <-ch:
			if !ok {
				closed = true
			}
		case <-timeout:
			t.Fatal("таймаут ожидания закрытия канала")
		}
	}

	assert.True(t, closed, "канал должен быть закрыт после отмены контекста")
}

func TestGeneratorRuntimeCounterMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := generatorRuntimeCounterMetrics(ctx)
	select {
	case m, ok := <-ch:
		assert.True(t, ok, "канал должен быть открыт")
		assert.Equal(t, "PollCount", m.ID)
		assert.Equal(t, types.CounterMetricType, m.Type)
		assert.NotNil(t, m.Delta)
		assert.Equal(t, int64(1), *m.Delta)
	case <-time.After(time.Second):
		t.Fatal("таймаут ожидания метрики")
	}
	cancel()
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "канал должен быть закрыт после отмены контекста")
	case <-time.After(time.Second):
		t.Fatal("таймаут ожидания закрытия канала")
	}
}

func TestGeneratorGoputilMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := generatorGoputilMetrics(ctx)
	received := map[string]bool{}
	timeout := time.After(2 * time.Second)
loop:
	for {
		select {
		case m, ok := <-ch:
			if !ok {
				break loop
			}
			assert.NotNil(t, m.Value)
			assert.Equal(t, types.GaugeMetricType, m.Type)
			received[m.ID] = true
		case <-timeout:
			t.Fatal("таймаут ожидания метрик")
		}
	}
	assert.True(t, received["TotalMemory"], "должна быть метрика TotalMemory")
	assert.True(t, received["FreeMemory"], "должна быть метрика FreeMemory")
	foundCPU := false
	for id := range received {
		if len(id) > 14 && id[:14] == "CPUutilization" {
			foundCPU = true
			break
		}
	}
	assert.True(t, foundCPU, "должна быть хотя бы одна метрика CPUutilization")
}

func TestGeneratorGoputilMetrics_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := generatorGoputilMetrics(ctx)
	cancel()
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "канал должен быть закрыт после отмены контекста")
	case <-time.After(time.Second):
		t.Fatal("таймаут ожидания закрытия канала")
	}
}

func TestMetricsFanIn_MergesChannels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch1 := make(chan types.Metrics, 2)
	ch2 := make(chan types.Metrics, 2)
	m1 := types.Metrics{MetricID: types.MetricID{ID: "m1"}}
	m2 := types.Metrics{MetricID: types.MetricID{ID: "m2"}}
	m3 := types.Metrics{MetricID: types.MetricID{ID: "m3"}}
	ch1 <- m1
	ch1 <- m2
	close(ch1)
	ch2 <- m3
	close(ch2)
	resultCh := metricsFanIn(ctx, ch1, ch2)
	received := map[string]bool{}
	for m := range resultCh {
		received[m.ID] = true
	}
	assert.Len(t, received, 3, "должны получить 3 метрики из двух каналов")
	assert.Contains(t, received, "m1")
	assert.Contains(t, received, "m2")
	assert.Contains(t, received, "m3")
}

func TestMetricsFanIn_ClosesFinalChannelAfterInputClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan types.Metrics, 1)
	ch <- types.Metrics{MetricID: types.MetricID{ID: "test"}}
	close(ch)
	resultCh := metricsFanIn(ctx, ch)
	for range resultCh {
	}
	assert.True(t, true)
}

func TestMetricsFanIn_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch1 := make(chan types.Metrics)
	ch2 := make(chan types.Metrics)
	resultCh := metricsFanIn(ctx, ch1, ch2)
	cancel()
	select {
	case _, ok := <-resultCh:
		assert.False(t, ok, "канал должен быть закрыт после отмены контекста")
	case <-time.After(time.Second):
		t.Fatal("таймаут ожидания закрытия канала после отмены контекста")
	}
}

func TestMetricsHandler_BatchProcessing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inputCh := make(chan types.Metrics)
	batchSize := 2
	mockFacade.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, batch []types.Metrics) error {
			assert.Len(t, batch, batchSize)
			return nil
		}).Times(1)

	resultCh := metricsHandler(ctx, mockFacade, inputCh, batchSize)
	go func() {
		defer close(inputCh)
		inputCh <- types.Metrics{MetricID: types.MetricID{ID: "m1"}}
		inputCh <- types.Metrics{MetricID: types.MetricID{ID: "m2"}}
	}()
	results := []result{}
	for r := range resultCh {
		results = append(results, r)
	}
	assert.Len(t, results, 2)
	for _, r := range results {
		assert.NoError(t, r.err)
	}
}

func TestMetricsHandler_PartialBatchOnClose(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inputCh := make(chan types.Metrics)
	batchSize := 3
	mockFacade.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, batch []types.Metrics) error {
			assert.Len(t, batch, 2)
			return nil
		}).Times(1)
	resultCh := metricsHandler(ctx, mockFacade, inputCh, batchSize)
	go func() {
		defer close(inputCh)
		inputCh <- types.Metrics{MetricID: types.MetricID{ID: "m1"}}
		inputCh <- types.Metrics{MetricID: types.MetricID{ID: "m2"}}
	}()
	results := []result{}
	for r := range resultCh {
		results = append(results, r)
	}
	assert.Len(t, results, 2)
	for _, r := range results {
		assert.NoError(t, r.err)
	}
}

func TestMetricsHandler_ContextCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ctx, cancel := context.WithCancel(context.Background())
	inputCh := make(chan types.Metrics)
	mockFacade.EXPECT().Updates(gomock.Any(), gomock.Any()).Times(0)
	resultCh := metricsHandler(ctx, mockFacade, inputCh, 2)
	cancel()
	select {
	case _, ok := <-resultCh:
		assert.False(t, ok, "resultCh должен быть закрыт после отмены контекста")
	case <-time.After(time.Second):
		t.Fatal("таймаут ожидания закрытия resultCh после отмены контекста")
	}
}

func TestMetricsFanOut_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	mockSemaphore := NewMockSemaphore(ctrl)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	numWorkers := 3
	batchSize := 2
	inputCh := make(chan types.Metrics)
	for i := 0; i < numWorkers; i++ {
		mockSemaphore.EXPECT().Acquire(gomock.Any(), int64(1)).Return(nil).Times(1)
		mockSemaphore.EXPECT().Release(int64(1)).Times(1)
	}
	channels := metricsFanOut(ctx, mockFacade, mockSemaphore, numWorkers, batchSize, inputCh)
	require.Len(t, channels, numWorkers)
	for _, ch := range channels {
		assert.NotNil(t, ch)
	}
}

func TestPollMetrics_ForwardMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inputCh := make(chan types.Metrics)
	pollInterval := 10 * time.Millisecond
	outputCh := pollMetrics(ctx, pollInterval, inputCh)
	val1 := 42.0
	delta1 := int64(5)
	val2 := 99.9
	delta2 := int64(100)
	go func() {
		inputCh <- types.Metrics{
			MetricID: types.MetricID{
				ID:   "metric1",
				Type: types.GaugeMetricType,
			},
			Value: &val1,
		}
		inputCh <- types.Metrics{
			MetricID: types.MetricID{
				ID:   "metric2",
				Type: types.CounterMetricType,
			},
			Delta: &delta1,
		}
		inputCh <- types.Metrics{
			MetricID: types.MetricID{
				ID:   "metric3",
				Type: types.GaugeMetricType,
			},
			Value: &val2,
		}
		inputCh <- types.Metrics{
			MetricID: types.MetricID{
				ID:   "metric4",
				Type: types.CounterMetricType,
			},
			Delta: &delta2,
		}
		close(inputCh)
	}()
	var got []types.Metrics
	timeout := time.After(200 * time.Millisecond)

LOOP:
	for {
		select {
		case m, ok := <-outputCh:
			if !ok {
				break LOOP
			}
			got = append(got, m)
		case <-timeout:
			t.Fatal("timeout waiting for metrics")
		}
	}
	assert.Len(t, got, 4)
	assert.Equal(t, "metric1", got[0].ID)
	assert.NotNil(t, got[0].Value)
	assert.Nil(t, got[0].Delta)
	assert.Equal(t, "metric2", got[1].ID)
	assert.NotNil(t, got[1].Delta)
	assert.Nil(t, got[1].Value)
	assert.Equal(t, "metric3", got[2].ID)
	assert.Equal(t, val2, *got[2].Value)
	assert.Equal(t, "metric4", got[3].ID)
	assert.Equal(t, delta2, *got[3].Delta)
}

func TestPollMetrics_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	inputCh := make(chan types.Metrics)
	pollInterval := 50 * time.Millisecond
	outputCh := pollMetrics(ctx, pollInterval, inputCh)
	cancel()
	select {
	case _, ok := <-outputCh:
		assert.False(t, ok, "output channel should be closed after context cancel")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout: output channel not closed after cancel")
	}
}

func TestPollMetrics_InputChannelClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inputCh := make(chan types.Metrics)
	pollInterval := 10 * time.Millisecond
	outputCh := pollMetrics(ctx, pollInterval, inputCh)
	close(inputCh)
	select {
	case _, ok := <-outputCh:
		assert.False(t, ok, "output channel should be closed when input channel is closed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout: output channel not closed after input channel close")
	}
}

func TestReportMetrics_ForwardMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inputCh := make(chan types.Metrics)
	reportInterval := 10 * time.Millisecond
	outputCh := reportMetrics(ctx, reportInterval, inputCh)
	val := 1.23
	delta := int64(42)
	go func() {
		inputCh <- types.Metrics{
			MetricID: types.MetricID{
				ID:   "gaugeMetric",
				Type: types.GaugeMetricType,
			},
			Value: &val,
		}
		inputCh <- types.Metrics{
			MetricID: types.MetricID{
				ID:   "counterMetric",
				Type: types.CounterMetricType,
			},
			Delta: &delta,
		}
		close(inputCh)
	}()
	var got []types.Metrics
	timeout := time.After(200 * time.Millisecond)
LOOP:
	for {
		select {
		case m, ok := <-outputCh:
			if !ok {
				break LOOP
			}
			got = append(got, m)
		case <-timeout:
			t.Fatal("timeout waiting for metrics")
		}
	}
	assert.Len(t, got, 2)
	assert.Equal(t, "gaugeMetric", got[0].ID)
	assert.Equal(t, val, *got[0].Value)
	assert.Equal(t, "counterMetric", got[1].ID)
	assert.Equal(t, delta, *got[1].Delta)
}

func TestReportMetrics_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	inputCh := make(chan types.Metrics)
	outputCh := reportMetrics(ctx, 10*time.Millisecond, inputCh)
	cancel()
	select {
	case _, ok := <-outputCh:
		assert.False(t, ok, "output channel should be closed after context cancel")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout: output channel not closed after cancel")
	}
}

func TestReportMetrics_InputChannelClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inputCh := make(chan types.Metrics)
	outputCh := reportMetrics(ctx, 10*time.Millisecond, inputCh)
	close(inputCh)
	select {
	case _, ok := <-outputCh:
		assert.False(t, ok, "output channel should be closed after input channel close")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout: output channel not closed after input channel closed")
	}
}

func TestProcessResults_PropagatesErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err1 := errors.New("first error")
	err2 := errors.New("second error")
	resultCh1 := make(chan result, 1)
	resultCh2 := make(chan result, 1)
	resultCh1 <- result{
		data: types.Metrics{
			MetricID: types.MetricID{
				ID:   "metric1",
				Type: types.GaugeMetricType,
			},
		},
		err: err1,
	}
	resultCh2 <- result{
		data: types.Metrics{
			MetricID: types.MetricID{
				ID:   "metric2",
				Type: types.CounterMetricType,
			},
		},
		err: err2,
	}
	close(resultCh1)
	close(resultCh2)
	errCh := processResults(ctx, []chan result{resultCh1, resultCh2})
	var received []error
	for e := range errCh {
		received = append(received, e)
	}
	assert.Len(t, received, 2)
	assert.Contains(t, received, err1)
	assert.Contains(t, received, err2)
}

func TestProcessResults_NoError_NoSend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resultCh := make(chan result, 2)
	resultCh <- result{
		data: types.Metrics{
			MetricID: types.MetricID{
				ID:   "metric1",
				Type: types.GaugeMetricType,
			},
		},
		err: nil,
	}
	close(resultCh)
	errCh := processResults(ctx, []chan result{resultCh})
	select {
	case err, ok := <-errCh:
		if ok {
			t.Fatalf("unexpected error received: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout: errCh was not closed")
	}
}

func TestProcessResults_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	resultCh := make(chan result)
	errCh := processResults(ctx, []chan result{resultCh})
	cancel()
	select {
	case _, ok := <-errCh:
		if !ok {
			return
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout: errCh was not closed after context cancel")
	}
}
