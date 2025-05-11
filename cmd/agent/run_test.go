package main

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	flagServerURL = "http://localhost:8080"
	flagPollInterval = 1
	flagReportInterval = 1
	flagLogLevel = "info"
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT)
	go func() {
		err := run()
		assert.NoError(t, err, "Run should complete without error on SIGINT")
	}()
	time.Sleep(500 * time.Millisecond)
	stopChan <- syscall.SIGINT
	select {
	case <-stopChan:
		assert.True(t, true, "Received shutdown signal")
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not shut down within the expected time")
	}
}
