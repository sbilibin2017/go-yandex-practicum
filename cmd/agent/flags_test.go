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
		"-u", "/newupdate",
		"-p", "5",
		"-r", "15",
		"-l", "debug",
	}
	opts := parseFlags()
	assert.Equal(t, "http://localhost:9090", opts.ServerAddress)
	assert.Equal(t, "/newupdate", opts.ServerUpdateEndpoint)
	assert.Equal(t, 5, opts.PollInterval)
	assert.Equal(t, 15, opts.ReportInterval)
	assert.Equal(t, "debug", opts.LogLevel)
}

func TestParseFlagsWithEnvironmentVariables(t *testing.T) {
	resetFlags()
	os.Setenv(envAddress, "http://localhost:9090")
	os.Setenv(envUpdateEndpoint, "/newupdate")
	os.Setenv(envPollInterval, "5")
	os.Setenv(envReportInterval, "15")
	os.Setenv(envLogLevel, "debug")
	os.Args = []string{"test"}

	opts := parseFlags()
	assert.Equal(t, "http://localhost:9090", opts.ServerAddress)
	assert.Equal(t, "/newupdate", opts.ServerUpdateEndpoint)
	assert.Equal(t, 5, opts.PollInterval)
	assert.Equal(t, 15, opts.ReportInterval)
	assert.Equal(t, "debug", opts.LogLevel)

	os.Unsetenv(envAddress)
	os.Unsetenv(envUpdateEndpoint)
	os.Unsetenv(envPollInterval)
	os.Unsetenv(envReportInterval)
	os.Unsetenv(envLogLevel)
}

func TestParseFlagsWithEnvironmentVariablesAndFlags(t *testing.T) {
	resetFlags()
	os.Setenv(envAddress, "http://localhost:9090")
	os.Setenv(envUpdateEndpoint, "/newupdate")
	os.Setenv(envPollInterval, "5")
	os.Setenv(envReportInterval, "15")
	os.Setenv(envLogLevel, "debug")

	os.Args = []string{
		"test",
		"-a", "http://localhost:8080",
		"-u", "/differentupdate",
		"-p", "10",
		"-r", "20",
		"-l", "info",
	}

	opts := parseFlags()
	// Переменные окружения имеют приоритет
	assert.Equal(t, "http://localhost:9090", opts.ServerAddress)
	assert.Equal(t, "/newupdate", opts.ServerUpdateEndpoint)
	assert.Equal(t, 5, opts.PollInterval)
	assert.Equal(t, 15, opts.ReportInterval)
	assert.Equal(t, "debug", opts.LogLevel)

	os.Unsetenv(envAddress)
	os.Unsetenv(envUpdateEndpoint)
	os.Unsetenv(envPollInterval)
	os.Unsetenv(envReportInterval)
	os.Unsetenv(envLogLevel)
}
