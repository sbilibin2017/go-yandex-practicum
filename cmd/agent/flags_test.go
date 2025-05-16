package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestParseFlagsWithFlags(t *testing.T) {
	resetFlags()
	os.Args = []string{
		"test",
		"-a", "http://localhost:9090",
		"-p", "5",
		"-r", "15",
		"-ll", "debug",
		"-k", "secretkey",
		"-l", "100",
		"-w", "20",
	}

	opts := parseFlags()
	assert.Equal(t, "http://localhost:9090", opts.ServerAddress)
	assert.Equal(t, 5, opts.PollInterval)
	assert.Equal(t, 15, opts.ReportInterval)
	assert.Equal(t, "debug", opts.LogLevel)
	assert.Equal(t, "secretkey", opts.Key)
	assert.Equal(t, 100, opts.RateLimit)
	assert.Equal(t, 20, opts.NumWorkers)
}

func TestParseFlagsWithEnvironmentVariables(t *testing.T) {
	resetFlags()

	os.Setenv(envAddress, "http://localhost:9090")
	os.Setenv(envPollInterval, "5")
	os.Setenv(envReportInterval, "15")
	os.Setenv(envLogLevel, "debug")
	os.Setenv(envKey, "envsecret")
	os.Setenv(envRateLimit, "200")
	os.Setenv(envNumWorkers, "30")

	os.Args = []string{"test"}

	opts := parseFlags()
	assert.Equal(t, "http://localhost:9090", opts.ServerAddress)
	assert.Equal(t, 5, opts.PollInterval)
	assert.Equal(t, 15, opts.ReportInterval)
	assert.Equal(t, "debug", opts.LogLevel)
	assert.Equal(t, "envsecret", opts.Key)
	assert.Equal(t, 200, opts.RateLimit)
	assert.Equal(t, 30, opts.NumWorkers)

	os.Unsetenv(envAddress)
	os.Unsetenv(envPollInterval)
	os.Unsetenv(envReportInterval)
	os.Unsetenv(envLogLevel)
	os.Unsetenv(envKey)
	os.Unsetenv(envRateLimit)
	os.Unsetenv(envNumWorkers)
}

func TestParseFlagsWithEnvironmentVariablesAndFlags(t *testing.T) {
	resetFlags()

	os.Setenv(envAddress, "http://localhost:9090")
	os.Setenv(envPollInterval, "5")
	os.Setenv(envReportInterval, "15")
	os.Setenv(envLogLevel, "debug")
	os.Setenv(envKey, "envsecret")
	os.Setenv(envRateLimit, "300")
	os.Setenv(envNumWorkers, "40")

	os.Args = []string{
		"test",
		"-a", "http://localhost:8080",
		"-p", "10",
		"-r", "20",
		"-ll", "info",
		"-k", "flagsecret",
		"-l", "400",
		"-w", "50",
	}

	opts := parseFlags()

	// Значения берутся из переменных окружения, а не из флагов
	assert.Equal(t, "http://localhost:9090", opts.ServerAddress)
	assert.Equal(t, 5, opts.PollInterval)
	assert.Equal(t, 15, opts.ReportInterval)
	assert.Equal(t, "debug", opts.LogLevel)
	assert.Equal(t, "envsecret", opts.Key)
	assert.Equal(t, 300, opts.RateLimit)
	assert.Equal(t, 40, opts.NumWorkers)

	os.Unsetenv(envAddress)
	os.Unsetenv(envPollInterval)
	os.Unsetenv(envReportInterval)
	os.Unsetenv(envLogLevel)
	os.Unsetenv(envKey)
	os.Unsetenv(envRateLimit)
	os.Unsetenv(envNumWorkers)
}
