package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupFlags() {
	os.Clearenv()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{"cmd"}
}

func TestDefaultFlags(t *testing.T) {
	setupFlags()
	parseFlags()
	assert.Equal(t, "http://localhost:8080", flagServerAddress, "Default server URL should be 'http://localhost:8080'")
	assert.Equal(t, "/update/", flagServerUpdateEndpoint, "Default update endpoint should be '/update/'")
	assert.Equal(t, 2, flagPollInterval, "Default poll interval should be 2 seconds")
	assert.Equal(t, 10, flagReportInterval, "Default report interval should be 10 seconds")
	assert.Equal(t, "info", flagLogLevel, "Default log level should be 'info'")
}

func TestFlagsFromCommandLine(t *testing.T) {
	setupFlags()
	os.Args = []string{"cmd", "-a", "http://example.com", "-u", "/custom-update", "-p", "5", "-r", "20", "-l", "debug"}
	parseFlags()
	assert.Equal(t, "http://example.com", flagServerAddress, "Flag -a should set the server URL")
	assert.Equal(t, "/custom-update", flagServerUpdateEndpoint, "Flag -u should set the server update endpoint")
	assert.Equal(t, 5, flagPollInterval, "Flag -p should set the poll interval")
	assert.Equal(t, 20, flagReportInterval, "Flag -r should set the report interval")
	assert.Equal(t, "debug", flagLogLevel, "Flag -l should set the log level")
}

func TestFlagsFromEnvironmentVariables(t *testing.T) {
	setupFlags()
	os.Setenv("ADDRESS", "http://env-server.com")
	os.Setenv("UPDATE_ENDPOINT", "/env-update")
	os.Setenv("POLL_INTERVAL", "3")
	os.Setenv("REPORT_INTERVAL", "15")
	os.Setenv("LOG_LEVEL", "warn")
	parseFlags()
	assert.Equal(t, "http://env-server.com", flagServerAddress, "Environment variable ADDRESS should override the flag")
	assert.Equal(t, "/env-update", flagServerUpdateEndpoint, "Environment variable UPDATE_ENDPOINT should override the flag")
	assert.Equal(t, 3, flagPollInterval, "Environment variable POLL_INTERVAL should override the flag")
	assert.Equal(t, 15, flagReportInterval, "Environment variable REPORT_INTERVAL should override the flag")
	assert.Equal(t, "warn", flagLogLevel, "Environment variable LOG_LEVEL should override the flag")
}

func TestFlagsFromBothCommandLineAndEnvironmentVariables(t *testing.T) {
	setupFlags()
	os.Args = []string{"cmd", "-a", "http://cli.com", "-u", "/cli-update", "-p", "8", "-r", "25", "-l", "debug"}
	os.Setenv("ADDRESS", "http://env.com")
	os.Setenv("UPDATE_ENDPOINT", "/env-update")
	os.Setenv("POLL_INTERVAL", "4")
	os.Setenv("REPORT_INTERVAL", "12")
	os.Setenv("LOG_LEVEL", "error")
	parseFlags()
	assert.Equal(t, "http://env.com", flagServerAddress, "Environment variable ADDRESS should override command-line flag")
	assert.Equal(t, "/env-update", flagServerUpdateEndpoint, "Environment variable UPDATE_ENDPOINT should override command-line flag")
	assert.Equal(t, 4, flagPollInterval, "Environment variable POLL_INTERVAL should override command-line flag")
	assert.Equal(t, 12, flagReportInterval, "Environment variable REPORT_INTERVAL should override command-line flag")
	assert.Equal(t, "error", flagLogLevel, "Environment variable LOG_LEVEL should override command-line flag")
}
