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
		"-a", "127.0.0.1:9000",
		"-d", "user:pass@/dbname",
		"-i", "123",
		"-f", "/tmp/files",
		"-r",
		"-k", "mykey",
		"-l", "debug",
		"-h", "CustomHeader", // добавлено
	}

	os.Clearenv()

	parseFlags()

	assert.Equal(t, "127.0.0.1:9000", flagServerAddress)
	assert.Equal(t, "user:pass@/dbname", flagDatabaseDSN)
	assert.Equal(t, 123, flagStoreInterval)
	assert.Equal(t, "/tmp/files", flagFileStoragePath)
	assert.True(t, flagRestore)
	assert.Equal(t, "mykey", flagKey)
	assert.Equal(t, "debug", flagLogLevel)
	assert.Equal(t, "CustomHeader", flagHeader) // проверка
}

func TestParseFlags_EnvOverrides(t *testing.T) {
	resetFlags()

	os.Args = []string{"cmd"}

	os.Setenv("ADDRESS", "10.0.0.1:8081")
	os.Setenv("DATABASE_DSN", "envuser:envpass@/envdb")
	os.Setenv("STORE_INTERVAL", "456")
	os.Setenv("FILE_STORAGE_PATH", "/env/files")
	os.Setenv("RESTORE", "true")
	os.Setenv("KEY", "envkey")
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("HEADER", "EnvHeader") // добавлено

	parseFlags()

	assert.Equal(t, "10.0.0.1:8081", flagServerAddress)
	assert.Equal(t, "envuser:envpass@/envdb", flagDatabaseDSN)
	assert.Equal(t, 456, flagStoreInterval)
	assert.Equal(t, "/env/files", flagFileStoragePath)
	assert.True(t, flagRestore)
	assert.Equal(t, "envkey", flagKey)
	assert.Equal(t, "error", flagLogLevel)
	assert.Equal(t, "EnvHeader", flagHeader) // проверка
}

func TestParseFlags_Defaults(t *testing.T) {
	resetFlags()

	os.Args = []string{"cmd"}
	os.Clearenv()

	parseFlags()

	assert.Equal(t, ":8080", flagServerAddress)
	assert.Equal(t, "", flagDatabaseDSN)
	assert.Equal(t, 300, flagStoreInterval)
	assert.Equal(t, "", flagFileStoragePath)
	assert.False(t, flagRestore)
	assert.Equal(t, "", flagKey)
	assert.Equal(t, "info", flagLogLevel)
	assert.Equal(t, "HashSHA256", flagHeader) // проверка дефолта
}
