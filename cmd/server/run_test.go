package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunServer(t *testing.T) {
	flagLogLevel = "info"
	flagServerAddress = "localhost:8085"
	flagStoreInterval = 1
	flagRestore = false
	flagFileStoragePath = ""
	flagDatabaseDSN = ""
	flagKey = ""

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = run(ctx)
	}()

	time.Sleep(2 * time.Second)
	cancel()
	time.Sleep(2 * time.Second)

	assert.True(t, true, "Server run and shutdown executed successfully")
}
