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

func clearEnvVars() {
	os.Unsetenv(envServerAddress)
	os.Unsetenv(envLogLevel)
	os.Unsetenv(envStoreInterval)
	os.Unsetenv(envFileStoragePath)
	os.Unsetenv(envRestore)
	os.Unsetenv(envDatabaseDSN)
}

func TestParseFlagsWithFlags(t *testing.T) {
	resetFlags()
	clearEnvVars()
	os.Args = []string{
		"test",
		"-d", "postgres://localhost/db",
		"-a", ":9090",
		"-l", "debug",
		"-i", "600",
		"-f", "/tmp/teststorage",
		"-r", "true",
	}
	result := parseFlags()
	assert.Equal(t, ":9090", result.ServerAddress)
	assert.Equal(t, "postgres://localhost/db", result.DatabaseDSN)
	assert.Equal(t, "debug", result.LogLevel)
	assert.Equal(t, 600, result.StoreInterval)
	assert.Equal(t, "/tmp/teststorage", result.FileStoragePath)
	assert.True(t, result.Restore)
}

func TestParseFlagsWithEnvironmentVariables(t *testing.T) {
	resetFlags()
	clearEnvVars()
	os.Setenv(envServerAddress, ":9090")
	os.Setenv(envLogLevel, "debug")
	os.Setenv(envStoreInterval, "600")
	os.Setenv(envFileStoragePath, "/tmp/teststorage")
	os.Setenv(envRestore, "true")
	os.Setenv(envDatabaseDSN, "postgres://localhost/db")
	os.Args = []string{"test"}

	result := parseFlags()
	assert.Equal(t, ":9090", result.ServerAddress)
	assert.Equal(t, "debug", result.LogLevel)
	assert.Equal(t, 600, result.StoreInterval)
	assert.Equal(t, "/tmp/teststorage", result.FileStoragePath)
	assert.True(t, result.Restore)
	assert.Equal(t, "postgres://localhost/db", result.DatabaseDSN)

	clearEnvVars()
}

func TestParseFlagsWithEnvironmentVariablesAndFlags(t *testing.T) {
	resetFlags()
	clearEnvVars()
	os.Setenv(envServerAddress, ":9090")
	os.Setenv(envLogLevel, "debug")
	os.Setenv(envStoreInterval, "600")
	os.Setenv(envFileStoragePath, "/tmp/teststorage")
	os.Setenv(envRestore, "true")
	os.Setenv(envDatabaseDSN, "postgres://localhost/db")
	os.Args = []string{
		"test",
		"-a", ":8080",
		"-l", "info",
		"-i", "300",
		"-f", "/tmp/storage",
		"-r", "false",
		"-d", "postgres://localhost/newdb",
	}

	result := parseFlags()
	assert.Equal(t, ":9090", result.ServerAddress)
	assert.Equal(t, "debug", result.LogLevel)
	assert.Equal(t, 600, result.StoreInterval)
	assert.Equal(t, "/tmp/teststorage", result.FileStoragePath)
	assert.True(t, result.Restore)
	assert.Equal(t, "postgres://localhost/db", result.DatabaseDSN)

	clearEnvVars()
}
