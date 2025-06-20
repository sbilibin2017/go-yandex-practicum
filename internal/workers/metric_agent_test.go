package workers

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRuntimeGaugeMetrics_Count(t *testing.T) {
	metrics := getRuntimeGaugeMetrics()
	assert.Len(t, metrics, 28, "expected exactly 28 gauge metrics")
}

func TestGetRuntimeCounterMetrics_Count(t *testing.T) {
	metrics := getRuntimeCounterMetrics()
	assert.Len(t, metrics, 1, "expected exactly 1 counter metric")
}

func TestGetGoputilMetrics_Count(t *testing.T) {
	ctx := context.Background()
	metrics := getGoputilMetrics(ctx)

	// At least TotalMemory + FreeMemory + CPU cores metrics
	expectedMin := 2 + runtime.NumCPU()
	assert.GreaterOrEqual(t, len(metrics), expectedMin, "expected at least %d goputil metrics", expectedMin)
}

func TestGeneratorMetrics(t *testing.T) {
	ctx := context.Background()

	input := []types.Metrics{
		{ID: "metric1"},
		{ID: "metric2"},
		{ID: "metric3"},
	}

	outCh := generatorMetrics(ctx, input)

	var got []types.Metrics
	for m := range outCh {
		got = append(got, m)
	}

	assert.Equal(t, len(input), len(got), "should output all input metrics")
	for i, metric := range input {
		assert.Equal(t, metric.ID, got[i].ID, "metric ID should match at index %d", i)
	}
}

func TestGeneratorMetrics_ContextDone(t *testing.T) {
	// Create a context that is already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	input := []types.Metrics{
		{ID: "metric1"},
		{ID: "metric2"},
	}

	outCh := generatorMetrics(ctx, input)

	// The output channel should be closed immediately and produce no metrics
	var got []types.Metrics
	for m := range outCh {
		got = append(got, m)
	}

	assert.Empty(t, got, "output channel should be closed without sending metrics when context is done")
}

func TestFanInMetrics_MergesChannels(t *testing.T) {
	ctx := context.Background()

	ch1 := make(chan types.Metrics, 2)
	ch2 := make(chan types.Metrics, 2)

	m1 := types.Metrics{ID: "metric1"}
	m2 := types.Metrics{ID: "metric2"}
	m3 := types.Metrics{ID: "metric3"}
	m4 := types.Metrics{ID: "metric4"}

	ch1 <- m1
	ch1 <- m2
	close(ch1)

	ch2 <- m3
	ch2 <- m4
	close(ch2)

	out := fanInMetrics(ctx, ch1, ch2)

	var results []types.Metrics
	for metric := range out {
		results = append(results, metric)
	}

	// We expect all four metrics to be present, order not guaranteed
	assert.ElementsMatch(t, []types.Metrics{m1, m2, m3, m4}, results)
}

func TestFanInMetrics_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ch1 := make(chan types.Metrics)
	ch2 := make(chan types.Metrics)

	out := fanInMetrics(ctx, ch1, ch2)

	// Cancel context to force early exit
	cancel()

	// Because context is canceled, the output channel should close quickly
	done := make(chan struct{})
	go func() {
		for range out {
			// drain any possible data
		}
		close(done)
	}()

	select {
	case <-done:
		// success: output channel closed
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for fanInMetrics output channel to close after context cancel")
	}
}

func TestWorkerMetricsUpdate_BatchesAndUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricUpdater(ctrl)

	ctx := context.Background()

	// Prepare metrics to send
	metrics := []types.Metrics{
		{ID: "m1"},
		{ID: "m2"},
		{ID: "m3"},
	}

	jobs := make(chan types.Metrics, len(metrics))

	// batchSize 2 for testing batch behaviour
	batchSize := 2

	// Expect Updates to be called with the first batch (2 metrics)
	mockUpdater.EXPECT().
		Updates(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, batch []*types.Metrics) error {
			assert.Len(t, batch, 2)
			assert.Equal(t, "m1", batch[0].ID)
			assert.Equal(t, "m2", batch[1].ID)
			return nil
		})

	// Expect Updates to be called with the last batch (1 metric)
	mockUpdater.EXPECT().
		Updates(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, batch []*types.Metrics) error {
			assert.Len(t, batch, 1)
			assert.Equal(t, "m3", batch[0].ID)
			return nil
		})

	// Send metrics into jobs channel
	for _, m := range metrics {
		jobs <- m
	}
	close(jobs)

	resultsCh := workerMetricsUpdate(ctx, mockUpdater, jobs, batchSize)

	var results []result
	for r := range resultsCh {
		results = append(results, r)
	}

	assert.Len(t, results, 2)

	// Check that results correspond to the batches sizes
	assert.Len(t, results[0].Data, 2)
	assert.NoError(t, results[0].Err)

	assert.Len(t, results[1].Data, 1)
	assert.NoError(t, results[1].Err)
}

func TestWorkerMetricsUpdate_ContextCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricUpdater(ctrl)

	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan types.Metrics)
	batchSize := 2

	resultsCh := workerMetricsUpdate(ctx, mockUpdater, jobs, batchSize)

	// Cancel context before sending anything
	cancel()

	// results channel should close quickly without any results
	done := make(chan struct{})
	go func() {
		for range resultsCh {
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for results channel to close after context cancel")
	}
}

func TestWorkerPoolMetricsUpdate_ProcessAllJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricUpdater(ctrl)
	ctx := context.Background()

	// Create jobs channel with some metrics
	jobs := make(chan types.Metrics, 5)
	metrics := []types.Metrics{
		{ID: "m1"},
		{ID: "m2"},
		{ID: "m3"},
		{ID: "m4"},
		{ID: "m5"},
	}

	for _, m := range metrics {
		jobs <- m
	}
	close(jobs)

	// We expect Updates to be called at least once
	// Because batchSize is 2, expect 3 calls: 2+2+1 metrics batches
	callCount := 0
	mockUpdater.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, batch []*types.Metrics) error {
			callCount++
			// batch size should not exceed 2
			assert.LessOrEqual(t, len(batch), 2)
			return nil
		}).
		Times(3)

	batchSize := 2
	rateLimit := 2

	resultsCh := workerPoolMetricsUpdate(ctx, mockUpdater, jobs, batchSize, rateLimit)

	// Collect results from the results channel
	var results []result
	for r := range resultsCh {
		results = append(results, r)
	}

	assert.Equal(t, 3, callCount)    // Confirm Updates was called 3 times
	assert.Equal(t, 3, len(results)) // Confirm we got 3 result batches
}

func TestWorkerPoolMetricsUpdate_ContextCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricUpdater(ctrl)
	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan types.Metrics)
	batchSize := 2
	rateLimit := 2

	resultsCh := workerPoolMetricsUpdate(ctx, mockUpdater, jobs, batchSize, rateLimit)

	// Cancel the context immediately
	cancel()

	done := make(chan struct{})
	go func() {
		for range resultsCh {
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for results channel to close after context cancel")
	}
}

func TestLogResults_ContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := logResults(ctx, make(chan result))
	assert.ErrorIs(t, err, context.Canceled)
}

func TestLogResults_ChannelClosed(t *testing.T) {
	ctx := context.Background()

	results := make(chan result)
	close(results) // close immediately

	err := logResults(ctx, results)
	assert.Nil(t, err)
}

func TestLogResults_ProcessResults(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := make(chan result, 2)

	results <- result{
		Data: []*types.Metrics{{ID: "test_metric"}},
		Err:  nil,
	}
	results <- result{
		Data: []*types.Metrics{{ID: "test_metric_err"}},
		Err:  errors.New("update failed"),
	}
	close(results)

	err := logResults(ctx, results)
	assert.Nil(t, err)
}

func TestStartMetricsReporting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricUpdater(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan types.Metrics, 10)
	batchSize := 2
	rateLimit := 1
	reportInterval := 1 // 1 second for test speed

	// Prepare dummy metrics
	metric := types.Metrics{ID: "test_metric"}

	// Expect Update to be called, simulate success (no error)
	mockUpdater.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	// Start the reporting function
	resultsCh := startMetricsReporting(ctx, mockUpdater, reportInterval, in, batchSize, rateLimit)

	// Send metrics to input channel
	in <- metric
	in <- metric

	// Wait some time to let ticker trigger workerPoolMetricsUpdate
	time.Sleep(1500 * time.Millisecond)

	// Cancel context to stop the goroutine
	cancel()

	// Collect results from resultsCh
	var results []result
	for res := range resultsCh {
		results = append(results, res)
	}

	// Assert we got some results back (depends on batchSize/rateLimit)
	assert.NotEmpty(t, results)
}

func TestStartMetricsPolling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pollInterval := 1 // 1 second

	metricsCh := startMetricsPolling(ctx, pollInterval)

	collected := make([]types.Metrics, 0)

	// Read some metrics for up to 3 seconds or until context is done
	for {
		select {
		case <-ctx.Done():
			goto done
		case metric, ok := <-metricsCh:
			if !ok {
				goto done
			}
			collected = append(collected, metric)
			// break early if we collected enough to be sure function works
			if len(collected) > 5 {
				goto done
			}
		}
	}

done:
	// We expect to have some metrics collected
	assert.NotEmpty(t, collected, "Expected some metrics to be collected")
}

func TestStartMetricAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricUpdater(ctrl)

	// Expect Updates to be called at least once with any metrics batch, return nil error
	mockUpdater.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	// Create a context that cancels after 500ms to end the test quickly
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Run the startMetricAgent with small intervals to trigger work quickly
	err := startMetricAgent(ctx, mockUpdater, 1, 1, 5, 2)

	// Because context times out, expect context.DeadlineExceeded or context.Canceled error
	assert.Error(t, err)
	assert.True(t, err == context.DeadlineExceeded || err == context.Canceled)
}

func TestNewMetricAgentWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricUpdater(ctrl)
	mockUpdater.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	worker := NewMetricAgentWorker(mockUpdater, 1, 1, 10, 2)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- worker(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		require.True(t, err == nil || errors.Is(err, context.Canceled),
			"expected nil or context.Canceled error, got %v", err)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("worker did not finish before timeout")
	}
}
