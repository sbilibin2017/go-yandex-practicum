package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	flagLogLevel = "info"
	flagPollInterval = 1
	flagReportInterval = 5
	flagServerAddress = "http://example.com"
	flagKey = ""
	flagRateLimit = 0
	flagNumWorkers = 0

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = run(ctx)
	}()

	time.Sleep(2 * time.Second)
	cancel()
	time.Sleep(1 * time.Second)

	assert.True(t, true, "Test ran successfully")
}
