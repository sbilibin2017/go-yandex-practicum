package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Save original function to restore after test
var origStartMetricAgentFunc = startMetricAgentFunc

func TestRun_Success(t *testing.T) {
	startMetricAgentFunc = func(ctx context.Context, _ string, _ string, _ string, _ string, _ int, _ int, _ int, _ int) error {
		<-ctx.Done()
		return nil
	}
	defer func() { startMetricAgentFunc = origStartMetricAgentFunc }()

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := run(ctx)
		assert.NoError(t, err)
	}()

	// Give run some time to start
	time.Sleep(50 * time.Millisecond)
	cancel() // simulate shutdown by canceling context

	wg.Wait()
}

func TestRun_ErrorFromAgent(t *testing.T) {
	wantErr := errors.New("agent failed")

	startMetricAgentFunc = func(ctx context.Context, _ string, _ string, _ string, _ string, _ int, _ int, _ int, _ int) error {
		return wantErr
	}
	defer func() { startMetricAgentFunc = origStartMetricAgentFunc }()

	ctx := context.Background()

	err := run(ctx)
	assert.Equal(t, wantErr, err)
}

func TestRun_GracefulShutdownWithoutAgentError(t *testing.T) {
	startMetricAgentFunc = func(ctx context.Context, _ string, _ string, _ string, _ string, _ int, _ int, _ int, _ int) error {
		<-ctx.Done()
		return nil
	}
	defer func() { startMetricAgentFunc = origStartMetricAgentFunc }()

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := run(ctx)
		assert.NoError(t, err)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel() // simulate graceful shutdown

	wg.Wait()
}
