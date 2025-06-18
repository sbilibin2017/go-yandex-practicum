package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/workers"
)

func TestRun_Shutdown(t *testing.T) {
	// Mock the worker function with the exact signature
	startMetricServerWorkerFunc = func(
		ctx context.Context,
		memoryListAll workers.MetricListAllRepository,
		memorySave workers.MetricSaveRepository,
		fileListAll workers.MetricListAllFileRepository,
		fileSave workers.MetricSaveFileRepository,
		restore bool,
		storeInterval int,
	) error {
		return nil
	}
	defer func() {
		startMetricServerWorkerFunc = workers.StartMetricServerWorker
	}()

	// Mock the HTTP server start to immediately return ErrServerClosed (graceful shutdown)
	startHTTPServerFunc = func(server *http.Server, errChan chan<- error) {
		errChan <- http.ErrServerClosed
	}
	defer func() {
		startHTTPServerFunc = func(server *http.Server, errChan chan<- error) {
			err := server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				errChan <- err
			}
		}
	}()

	// Create a context that is immediately cancelled to simulate shutdown signal
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Run should return nil without error on shutdown
	if err := run(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
