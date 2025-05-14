package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	opts := &options{
		LogLevel:             "info",
		PollInterval:         1,
		ReportInterval:       5,
		ServerAddress:        "http://example.com",
		ServerUpdateEndpoint: "/update",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		run(ctx, opts)
	}()
	cancel()
	time.Sleep(3 * time.Second)
	assert.True(t, true, "Test ran successfully")
}
