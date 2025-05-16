package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestParseFlags_CommandLineArgs(t *testing.T) {
	resetFlags()

	os.Args = []string{
		"cmd",
		"-a", "http://127.0.0.1:9000",
		"-p", "15",
		"-r", "30",
		"-ll", "debug",
		"-k", "mykey",
		"-l", "50",
		"-w", "7",
	}

	os.Clearenv()

	parseFlags()

	assert.Equal(t, "http://127.0.0.1:9000", flagServerAddress)
	assert.Equal(t, 15, flagPollInterval)
	assert.Equal(t, 30, flagReportInterval)
	assert.Equal(t, "debug", flagLogLevel)
	assert.Equal(t, "mykey", flagKey)
	assert.Equal(t, 50, flagRateLimit)
	assert.Equal(t, 7, flagNumWorkers)
}

func TestParseFlags_EnvOverrides(t *testing.T) {
	resetFlags()

	os.Args = []string{"cmd"}

	os.Setenv("ADDRESS", "http://10.0.0.1:8081")
	os.Setenv("POLL_INTERVAL", "20")
	os.Setenv("REPORT_INTERVAL", "40")
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("KEY", "envkey")
	os.Setenv("RATE_LIMIT", "100")
	os.Setenv("NUM_WORKERS", "12")

	parseFlags()

	assert.Equal(t, "http://10.0.0.1:8081", flagServerAddress)
	assert.Equal(t, 20, flagPollInterval)
	assert.Equal(t, 40, flagReportInterval)
	assert.Equal(t, "error", flagLogLevel)
	assert.Equal(t, "envkey", flagKey)
	assert.Equal(t, 100, flagRateLimit)
	assert.Equal(t, 12, flagNumWorkers)
}

func TestParseFlags_Defaults(t *testing.T) {
	resetFlags()

	os.Args = []string{"cmd"}
	os.Clearenv()

	parseFlags()

	assert.Equal(t, "http://localhost:8080", flagServerAddress)
	assert.Equal(t, 2, flagPollInterval)
	assert.Equal(t, 10, flagReportInterval)
	assert.Equal(t, "info", flagLogLevel)
	assert.Equal(t, "", flagKey)
	assert.Equal(t, 0, flagRateLimit)
	assert.Equal(t, 5, flagNumWorkers)
}
