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
	flag.CommandLine.Parse([]string{})
}

func TestDefaultFlags(t *testing.T) {
	setupFlags()
	parseFlags()
	assert.Equal(t, ":8080", flagRunAddress, "Default server address should be ':8080'")
	assert.Equal(t, "info", flagLogLevel, "Default log level should be 'info'")
}

func TestFlagsFromCommandLine(t *testing.T) {
	setupFlags()
	os.Args = []string{"cmd", "-a", ":9090", "-l", "debug"}
	parseFlags()
	assert.Equal(t, ":9090", flagRunAddress, "Flag -a should set the server address")
	assert.Equal(t, "debug", flagLogLevel, "Flag -l should set the log level")
}

func TestFlagsFromEnvironmentVariables(t *testing.T) {
	setupFlags()
	os.Setenv("ADDRESS", ":5000")
	os.Setenv("LOG_LEVEL", "error")
	parseFlags()
	assert.Equal(t, ":5000", flagRunAddress, "Environment variable ADDRESS should override the flag")
	assert.Equal(t, "error", flagLogLevel, "Environment variable LOG_LEVEL should override the flag")
}

func TestFlagsFromBothCommandLineAndEnvironmentVariables(t *testing.T) {
	setupFlags()
	os.Args = []string{"cmd", "-a", ":9090", "-l", "debug"}
	os.Setenv("ADDRESS", ":5000")
	os.Setenv("LOG_LEVEL", "error")
	parseFlags()
	assert.Equal(t, ":5000", flagRunAddress, "Environment variable ADDRESS should override the flag from command line")
	assert.Equal(t, "error", flagLogLevel, "Environment variable LOG_LEVEL should override the flag from command line")
}
