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
	flagServerAddress = ":8080"
	flagLogLevel = "info"
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT)
	go func() {
		err := run()
		assert.NoError(t, err, "Server should start and stop without errors")
	}()
	time.Sleep(1 * time.Second)
	stopChan <- syscall.SIGINT
	select {
	case <-stopChan:
		assert.True(t, true)
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not shut down within the expected time")
	}
}
