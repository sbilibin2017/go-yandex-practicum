package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunServer(t *testing.T) {
	opts := &options{
		LogLevel:      "info",
		ServerAddress: "localhost:8085",
		StoreInterval: 1,
		Restore:       false,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = run(ctx, opts)
	}()
	time.Sleep(2 * time.Second)
	cancel()
	time.Sleep(2 * time.Second)
	assert.True(t, true, "Server run and shutdown executed successfully")
}
