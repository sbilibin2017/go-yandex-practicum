package workers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgconn"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestConsumeMetricsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ch := make(chan types.Metrics, 2)
	ch <- types.Metrics{
		MetricID: types.MetricID{ID: "Metric1", Type: types.GaugeMetricType},
		Value:    func() *float64 { v := 1.1; return &v }(),
	}
	ch <- types.Metrics{
		MetricID: types.MetricID{ID: "Metric2", Type: types.CounterMetricType},
		Delta:    func() *int64 { v := int64(10); return &v }(),
	}
	mockFacade.EXPECT().Updates(gomock.Any(), gomock.Len(2)).Return(fmt.Errorf("mock error"))
	consumeMetrics(context.Background(), mockFacade, ch)
	assert.Equal(t, 0, len(ch), "Expected all metrics consumed from channel")
}

func TestConsumeMetricsSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ch := make(chan types.Metrics, 2)
	ch <- types.Metrics{
		MetricID: types.MetricID{ID: "MetricA", Type: types.GaugeMetricType},
		Value:    func() *float64 { v := 3.14; return &v }(),
	}
	ch <- types.Metrics{
		MetricID: types.MetricID{ID: "MetricB", Type: types.CounterMetricType},
		Delta:    func() *int64 { v := int64(7); return &v }(),
	}
	mockFacade.EXPECT().Updates(gomock.Any(), gomock.Len(2)).Return(nil)
	consumeMetrics(context.Background(), mockFacade, ch)
	assert.Equal(t, 0, len(ch), "Expected all metrics consumed from channel")
}

func TestStartMetricAgentWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFacade := NewMockMetricFacade(ctrl)
	ch := make(chan types.Metrics, 100)
	mockFacade.EXPECT().Updates(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pollTicker := time.NewTicker(100 * time.Millisecond)
	reportTicker := time.NewTicker(200 * time.Millisecond)
	done := make(chan struct{})
	go func() {
		defer close(done)
		err := StartMetricAgentWorker(ctx, mockFacade, ch, pollTicker, reportTicker)
		assert.NoError(t, err)
	}()
	time.Sleep(500 * time.Millisecond)
	cancel()
	<-done
	assert.Greater(t, len(ch), 0, "Expected some metrics produced during polling")
}

func TestWithRetries_SuccessFirstTry(t *testing.T) {
	ctx := context.Background()
	calls := 0
	err := withRetries(ctx, func(ctx context.Context) error {
		calls++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestWithRetries_NonRetriableError(t *testing.T) {
	ctx := context.Background()
	calls := 0
	err := withRetries(ctx, func(ctx context.Context) error {
		calls++
		return errors.New("fatal error")
	})

	assert.EqualError(t, err, "fatal error")
	assert.Equal(t, 1, calls)
}

func TestWithRetries_RetriablePgError(t *testing.T) {
	ctx := context.Background()
	calls := 0
	err := withRetries(ctx, func(ctx context.Context) error {
		calls++
		return &pgconn.PgError{Code: "08006"}
	})

	assert.Error(t, err)
	assert.Equal(t, 4, calls)
}

func TestWithRetries_RetriableFileError(t *testing.T) {
	ctx := context.Background()
	calls := 0
	err := withRetries(ctx, func(ctx context.Context) error {
		calls++
		return &os.PathError{
			Op:   "open",
			Path: "/fake/path",
			Err:  syscall.EAGAIN,
		}
	})

	assert.Error(t, err)
	assert.Equal(t, 4, calls)
}
