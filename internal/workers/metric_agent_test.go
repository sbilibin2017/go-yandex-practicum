package workers

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/semaphore"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestProduceRuntimeGaugeMetricsStage_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan types.Metrics)
	out := produceRuntimeGaugeMetricsStage(ctx, in)
	go func() {
		in <- types.Metrics{}
		close(in)
	}()
	done := make(chan struct{})
	go func() {
		for range out {
		}
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for output channel to close after context cancel")
	}
}

func TestProduceRuntimeCounterMetricsStage_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan types.Metrics)
	out := produceRuntimeCounterMetricsStage(ctx, in)
	go func() {
		in <- types.Metrics{}
		close(in)
	}()
	done := make(chan struct{})
	go func() {
		for range out {
		}
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for output channel to close after context cancel")
	}
}

func TestProduceGopsutilGaugeMetricsStage_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan types.Metrics)
	out := produceGopsutilGaugeMetricsStage(ctx, in)
	go func() {
		in <- types.Metrics{}
		close(in)
	}()
	done := make(chan struct{})
	go func() {
		for range out {
		}
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for output channel to close after context cancel")
	}
}

func TestMetricsFanOut_MergesChannels(t *testing.T) {
	ctx := context.Background()
	ch1 := make(chan types.Metrics, 2)
	ch2 := make(chan types.Metrics, 2)
	ch1 <- types.Metrics{MetricID: types.MetricID{ID: "m1", Type: types.GaugeMetricType}}
	ch1 <- types.Metrics{MetricID: types.MetricID{ID: "m2", Type: types.GaugeMetricType}}
	close(ch1)
	ch2 <- types.Metrics{MetricID: types.MetricID{ID: "m3", Type: types.GaugeMetricType}}
	ch2 <- types.Metrics{MetricID: types.MetricID{ID: "m4", Type: types.GaugeMetricType}}
	close(ch2)
	out := metricsFanOut(ctx, ch1, ch2)
	var results []string
	for m := range out {
		results = append(results, m.ID)
	}
	assert.ElementsMatch(t, []string{"m1", "m2", "m3", "m4"}, results)
}

func TestMetricsFanOut_ClosesOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan types.Metrics)
	out := metricsFanOut(ctx, ch)
	cancel()
	select {
	case _, ok := <-out:
		assert.False(t, ok, "expected output channel to be closed after context cancel")
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for output channel to close after context cancel")
	}
}

func TestMetricsFanOut_StopsOnContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan types.Metrics, 1)
	out := metricsFanOut(ctx, ch)
	ch <- types.Metrics{MetricID: types.MetricID{ID: "m", Type: types.GaugeMetricType}}
	cancel()
	var results []string
	for m := range out {
		results = append(results, m.ID)
	}
	assert.True(t, len(results) <= 1)
	if len(results) == 1 {
		assert.Equal(t, "m", results[0])
	}
}

func TestProduceRuntimeGaugeMetricsStage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	in := make(chan types.Metrics, 1)
	in <- types.Metrics{}
	close(in)
	out := produceRuntimeGaugeMetricsStage(ctx, in)
	ids := make(map[string]bool)
	for m := range out {
		ids[m.MetricID.ID] = true
	}
	assert.Contains(t, ids, "Alloc")
	assert.Contains(t, ids, "RandomValue")
	assert.Contains(t, ids, "HeapAlloc")
}

func TestProduceRuntimeCounterMetricsStage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	in := make(chan types.Metrics, 2)
	in <- types.Metrics{}
	in <- types.Metrics{}
	close(in)
	out := produceRuntimeCounterMetricsStage(ctx, in)
	var count int
	for m := range out {
		assert.Equal(t, "PollCount", m.MetricID.ID)
		assert.Equal(t, types.CounterMetricType, m.MetricID.Type)
		count++
	}
	assert.Equal(t, 1, count)
}

func TestProduceGopsutilGaugeMetricsStage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	in := make(chan types.Metrics, 1)
	in <- types.Metrics{}
	close(in)
	out := produceGopsutilGaugeMetricsStage(ctx, in)
	metricsReceived := make(map[string]bool)
	for m := range out {
		metricsReceived[m.MetricID.ID] = true
	}
	assert.Contains(t, metricsReceived, "TotalMemory", "Expected TotalMemory metric")
	assert.Contains(t, metricsReceived, "FreeMemory", "Expected FreeMemory metric")
	foundCPU := false
	for id := range metricsReceived {
		if len(id) >= len("CPUutilization") && id[:len("CPUutilization")] == "CPUutilization" {
			foundCPU = true
			break
		}
	}
	assert.True(t, foundCPU, "Expected at least one CPUutilization metric")
}

func TestGenerator_ProducesMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pollTicker := time.NewTicker(10 * time.Millisecond)
	defer pollTicker.Stop()
	ch := generator(ctx, pollTicker)
	var received []types.Metrics
	timeout := time.After(100 * time.Millisecond)
loop:
	for {
		select {
		case m, ok := <-ch:
			if !ok {
				break loop
			}
			received = append(received, m)
			if len(received) > 10 {
				break loop
			}
		case <-timeout:
			break loop
		}
	}
	assert.Greater(t, len(received), 0, "expected some metrics to be generated")
}

func TestMetricsWorker_SendsBatches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockHandler := NewMockMetricFacade(ctrl)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reportTicker := time.NewTicker(50 * time.Millisecond)
	defer reportTicker.Stop()
	taskCh := make(chan types.Metrics, 10)
	mockHandler.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, metrics []types.Metrics) error {
			assert.NotEmpty(t, metrics)
			return nil
		}).
		AnyTimes()
	sem := semaphore.NewWeighted(1)
	resultCh := metricsWorker(ctx, mockHandler, taskCh, reportTicker, sem)
	for i := 0; i < 5; i++ {
		val := float64(i)
		taskCh <- types.Metrics{
			MetricID: types.MetricID{ID: "metric" + strconv.Itoa(i), Type: types.GaugeMetricType},
			Value:    &val,
		}
	}
	time.Sleep(60 * time.Millisecond)
	close(taskCh)
	var results []Result
	timeout := time.After(200 * time.Millisecond)
loop:
	for {
		select {
		case res, ok := <-resultCh:
			if !ok {
				break loop
			}
			results = append(results, res)
		case <-timeout:
			break loop
		}
	}
	require.Greater(t, len(results), 0)
	assert.Equal(t, 5, len(results))
	for _, res := range results {
		assert.NoError(t, res.err)
	}
}

func TestMetricsWorker_HandlerErrorIsReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockHandler := NewMockMetricFacade(ctrl)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reportTicker := time.NewTicker(50 * time.Millisecond)
	defer reportTicker.Stop()
	taskCh := make(chan types.Metrics, 10)
	expectedErr := assert.AnError
	mockHandler.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Return(expectedErr).
		Times(1)
	sem := semaphore.NewWeighted(1)
	resultCh := metricsWorker(ctx, mockHandler, taskCh, reportTicker, sem)
	val := float64(42)
	taskCh <- types.Metrics{
		MetricID: types.MetricID{ID: "failMetric", Type: types.GaugeMetricType},
		Value:    &val,
	}
	time.Sleep(60 * time.Millisecond)
	close(taskCh)
	var results []Result
	timeout := time.After(200 * time.Millisecond)
loop:
	for {
		select {
		case res, ok := <-resultCh:
			if !ok {
				break loop
			}
			results = append(results, res)
		case <-timeout:
			break loop
		}
	}
	require.Greater(t, len(results), 0)
	for _, res := range results {
		assert.Error(t, res.err)
		assert.Equal(t, expectedErr, res.err)
	}
}

func TestHandleResults_LogsOnResult(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan Result, 2)
	errCh := make(chan error, 1)

	handleResults(ctx, ch, errCh)

	val := float64(1)
	ch <- Result{
		data: types.Metrics{MetricID: types.MetricID{ID: "success", Type: types.GaugeMetricType}, Value: &val},
		err:  nil,
	}
	ch <- Result{
		data: types.Metrics{MetricID: types.MetricID{ID: "fail", Type: types.GaugeMetricType}, Value: &val},
		err:  assert.AnError,
	}
	close(ch)

	select {
	case err := <-errCh:
		require.Equal(t, assert.AnError, err)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ожидалась ошибка, но она не пришла в errCh")
	}
}

func TestStartMetricAgentWorker_RunsWithoutPanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pollInterval := 1
	reportInterval := 1
	numWorkers := 2
	mockHandler := NewMockMetricFacade(ctrl)
	mockHandler.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	worker := NewMetricAgentWorker(mockHandler, pollInterval, reportInterval, numWorkers)
	go worker.Start(ctx)
	time.Sleep(1500 * time.Millisecond)
	cancel()
	time.Sleep(100 * time.Millisecond)
}

func TestFanInResults_MergesChannels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch1 := make(chan Result, 2)
	ch2 := make(chan Result, 2)
	val1 := 1.23
	val2 := 4.56
	val3 := 7.89
	val4 := 0.12
	ch1 <- Result{data: types.Metrics{MetricID: types.MetricID{ID: "m1", Type: types.GaugeMetricType}, Value: &val1}, err: nil}
	ch1 <- Result{data: types.Metrics{MetricID: types.MetricID{ID: "m2", Type: types.GaugeMetricType}, Value: &val2}, err: nil}
	close(ch1)
	ch2 <- Result{data: types.Metrics{MetricID: types.MetricID{ID: "m3", Type: types.GaugeMetricType}, Value: &val3}, err: nil}
	ch2 <- Result{data: types.Metrics{MetricID: types.MetricID{ID: "m4", Type: types.GaugeMetricType}, Value: &val4}, err: nil}
	close(ch2)
	out := fanInResults(ctx, ch1, ch2)
	var results []string
	for r := range out {
		results = append(results, r.data.MetricID.ID)
	}
	assert.ElementsMatch(t, []string{"m1", "m2", "m3", "m4"}, results)
}

func TestFanInResults_ClosesOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan Result)
	out := fanInResults(ctx, ch)
	cancel()
	select {
	case _, ok := <-out:
		assert.False(t, ok, "expected output channel to be closed after context cancel")
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for output channel to close after context cancel")
	}
}
