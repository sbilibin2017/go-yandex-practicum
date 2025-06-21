package workers

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test default AgentWorkerConfig creation
func TestNewAgentWorkerConfig_Defaults(t *testing.T) {
	cfg := NewAgentWorkerConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, 0, cfg.pollInterval)
	assert.Equal(t, 0, cfg.reportInterval)
	assert.Equal(t, 0, cfg.batchSize)
	assert.Equal(t, 0, cfg.rateLimit)
	assert.Nil(t, cfg.updater)
}

// Test setting options for AgentWorkerConfig
func TestAgentWorkerConfig_WithOptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockUpdater(ctrl)

	cfg := NewAgentWorkerConfig(
		WithPollInterval(10),
		WithReportInterval(20),
		WithBatchSize(50),
		WithRateLimit(5),
		WithUpdater(mockUpdater),
	)

	assert.Equal(t, 10, cfg.pollInterval)
	assert.Equal(t, 20, cfg.reportInterval)
	assert.Equal(t, 50, cfg.batchSize)
	assert.Equal(t, 5, cfg.rateLimit)
	assert.Equal(t, mockUpdater, cfg.updater)
}

// Test running AgentWorker without errors (with mocked Updater)
func TestNewAgentWorker_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockUpdater(ctrl)
	mockUpdater.EXPECT().Updates(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	worker := NewAgentWorker(
		WithPollInterval(1),
		WithReportInterval(1),
		WithBatchSize(3),
		WithRateLimit(2),
		WithUpdater(mockUpdater),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := worker(ctx)
	assert.NoError(t, err)
}

// TestStartMetricsPolling_CancelImmediately tests cancel before receiving metrics
func TestStartMetricsPolling_CancelImmediately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	ch := startMetricsPolling(ctx, 1)
	assert.NotNil(t, ch)

	// The channel should close quickly due to canceled context
	_, ok := <-ch
	assert.False(t, ok, "channel should be closed due to canceled context")
}

func TestLogResults_StopsOnContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	resultsCh := make(chan Result)

	err := logResults(ctx, resultsCh)
	assert.NoError(t, err)

	// Cancel context to stop the goroutine inside logResults
	cancel()

	// Wait briefly to allow goroutine to exit cleanly
	time.Sleep(100 * time.Millisecond)
}

func TestLogResults_StopsWhenResultsChannelClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultsCh := make(chan Result)

	err := logResults(ctx, resultsCh)
	assert.NoError(t, err)

	// Close the results channel to trigger stop condition
	close(resultsCh)

	// Wait briefly to allow goroutine to exit cleanly
	time.Sleep(100 * time.Millisecond)
}

func TestLogResults_ProcessErrorAndSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultsCh := make(chan Result, 2)

	// Send a successful result
	resultsCh <- Result{Data: []*types.Metrics{{ID: "metric1"}}, Err: nil}

	// Send an error result
	resultsCh <- Result{Data: []*types.Metrics{{ID: "metric2"}}, Err: assert.AnError}

	close(resultsCh)

	err := logResults(ctx, resultsCh)
	assert.NoError(t, err)

	// Wait briefly to allow goroutine to process the channel
	time.Sleep(100 * time.Millisecond)
}

func TestStartMetricsPolling(t *testing.T) {
	// Use a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start polling with 1-second interval
	metricsCh := startMetricsPolling(ctx, 1)

	// Collect some metrics from the channel
	var collected []*types.Metrics
	timeout := time.After(3 * time.Second)

LOOP:
	for {
		select {
		case m, ok := <-metricsCh:
			if !ok {
				break LOOP
			}
			require.NotNil(t, m)
			collected = append(collected, m)

			// After collecting at least some metrics, break early
			if len(collected) >= 5 {
				break LOOP
			}
		case <-timeout:
			t.Fatal("timed out waiting for metrics")
		}
	}

	// Assert we received metrics
	assert.GreaterOrEqual(t, len(collected), 5, "should collect at least 5 metrics")

	// Check some metric fields have expected values
	for _, metric := range collected {
		assert.NotEmpty(t, metric.ID)
		assert.True(t, metric.MType == types.Gauge || metric.MType == types.Counter)
		assert.True(t, (metric.Value != nil) || (metric.Delta != nil))
	}

	// Cancel context and check channel is closed
	cancel()

	// Drain remaining metrics and verify channel closes
	for range metricsCh {
		// no-op
	}
}

func TestStartMetricsReporting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	float64Ptr := func(f float64) *float64 {
		return &f
	}

	mockUpdater := NewMockUpdater(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Prepare some sample metrics
	metrics := []*types.Metrics{
		{ID: "m1", MType: types.Gauge, Value: float64Ptr(1)},
		{ID: "m2", MType: types.Gauge, Value: float64Ptr(2)},
		{ID: "m3", MType: types.Gauge, Value: float64Ptr(3)},
		{ID: "m4", MType: types.Gauge, Value: float64Ptr(4)},
		{ID: "m5", MType: types.Gauge, Value: float64Ptr(5)},
	}

	inCh := make(chan *types.Metrics, len(metrics))
	for _, m := range metrics {
		inCh <- m
	}
	close(inCh)

	// We expect the updater to be called with batches of size 2
	batchSize := 2
	rateLimit := 2
	reportInterval := 1

	// Set expectations on mockUpdater.Updates to be called with batches of size <= batchSize
	// The test doesn't know the exact batches, so use DoAndReturn to validate the batch size.
	mockUpdater.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, batch []*types.Metrics) error {
			assert.LessOrEqual(t, len(batch), batchSize)
			for _, metric := range batch {
				assert.NotNil(t, metric)
				assert.NotEmpty(t, metric.ID)
			}
			return nil
		}).
		Times(3) // Because we have 5 metrics and batch size 2, expect 3 calls

	// Start the reporting
	resultsCh := startMetricsReporting(ctx, reportInterval, mockUpdater, inCh, batchSize, rateLimit)

	// Collect results, expect 3 results with no error
	var results []Result
	timeout := time.After(3 * time.Second)

	for len(results) < 3 {
		select {
		case res, ok := <-resultsCh:
			if !ok {
				break
			}
			results = append(results, res)
			assert.NoError(t, res.Err)
			assert.NotEmpty(t, res.Data)
		case <-timeout:
			t.Fatal("timeout waiting for reporting results")
		}
	}

	cancel() // cancel context to clean up goroutines

	// Drain the channel to allow goroutine exit
	for range resultsCh {
	}

	// Final assertions on results count
	assert.Len(t, results, 3)
}
