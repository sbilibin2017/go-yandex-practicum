package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/configs"
	"github.com/stretchr/testify/assert"
)

// helpers to get pointers for literals
func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }

func TestParseFlags_NoError(t *testing.T) {
	err := parseFlags()
	assert.NoError(t, err)
}

func TestParseEnvs(t *testing.T) {
	os.Setenv("ADDRESS", "localhost:9999")
	os.Setenv("DATABASE_DSN", "env_dsn")
	os.Setenv("STORE_INTERVAL", "123")
	os.Setenv("FILE_STORAGE_PATH", "/env/file")
	os.Setenv("RESTORE", "true")
	os.Setenv("KEY", "env_key")
	os.Setenv("CRYPTO_KEY", "/env/key.pem")
	os.Setenv("CONFIG", "/some/path")
	os.Setenv("TRUSTED_SUBNET", "192.168.0.0/16")

	defer func() {
		_ = os.Unsetenv("ADDRESS")
		_ = os.Unsetenv("DATABASE_DSN")
		_ = os.Unsetenv("STORE_INTERVAL")
		_ = os.Unsetenv("FILE_STORAGE_PATH")
		_ = os.Unsetenv("RESTORE")
		_ = os.Unsetenv("KEY")
		_ = os.Unsetenv("CRYPTO_KEY")
		_ = os.Unsetenv("CONFIG")
		_ = os.Unsetenv("TRUSTED_SUBNET")
	}()

	err := parseEnvs()
	assert.NoError(t, err)
	assert.Equal(t, "localhost:9999", flagServerAddress)
	assert.Equal(t, "env_dsn", flagDatabaseDSN)
	assert.Equal(t, 123, flagStoreInterval)
	assert.Equal(t, "/env/file", flagFileStoragePath)
	assert.True(t, flagRestore)
	assert.Equal(t, "env_key", flagKey)
	assert.Equal(t, "/env/key.pem", flagCryptoKey)
	assert.Equal(t, "/some/path", flagConfigPath)
	assert.Equal(t, "192.168.0.0/16", flagTrustedSubnet)
}

func TestParseConfigFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-config-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Now we use pointers for each field
	cfg := configs.ServerConfig{
		ServerAddress:   strPtr("127.0.0.1:9090"),
		DatabaseDSN:     strPtr("config_dsn"),
		StoreInterval:   intPtr(999),
		FileStoragePath: strPtr("/tmp/file"),
		Restore:         boolPtr(true),
		Key:             strPtr("config_key"),
		CryptoKey:       strPtr("/tmp/key.pem"),
		TrustedSubnet:   strPtr("10.0.0.0/8"),
	}

	err = json.NewEncoder(tmpFile).Encode(cfg)
	assert.NoError(t, err)
	_ = tmpFile.Close()

	flagConfigPath = tmpFile.Name()
	err = parseConfigFile()
	assert.NoError(t, err)

	assert.Equal(t, "127.0.0.1:9090", flagServerAddress)
	assert.Equal(t, "config_dsn", flagDatabaseDSN)
	assert.Equal(t, 999, flagStoreInterval)
	assert.Equal(t, "/tmp/file", flagFileStoragePath)
	assert.True(t, flagRestore)
	assert.Equal(t, "config_key", flagKey)
	assert.Equal(t, "/tmp/key.pem", flagCryptoKey)
	assert.Equal(t, "10.0.0.0/8", flagTrustedSubnet)
}

func TestFlags_Success(t *testing.T) {
	// Prepare environment variables
	os.Setenv("ADDRESS", "localhost:9999")
	os.Setenv("DATABASE_DSN", "env_dsn")
	os.Setenv("STORE_INTERVAL", "100")
	os.Setenv("FILE_STORAGE_PATH", "/env/path")
	os.Setenv("RESTORE", "true")
	os.Setenv("KEY", "env_key")
	os.Setenv("CRYPTO_KEY", "/env/crypto.key")
	os.Setenv("TRUSTED_SUBNET", "192.168.0.0/24")

	defer func() {
		os.Clearenv()
	}()

	// Create temporary config file
	tmpFile, err := os.CreateTemp("", "config-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	cfg := configs.ServerConfig{
		ServerAddress:   strPtr("127.0.0.1:9090"),
		DatabaseDSN:     strPtr("config_dsn"),
		StoreInterval:   intPtr(200),
		FileStoragePath: strPtr("/config/path"),
		Restore:         boolPtr(false),
		Key:             strPtr("config_key"),
		CryptoKey:       strPtr("/config/crypto.key"),
		TrustedSubnet:   strPtr("10.0.0.0/8"),
	}
	err = json.NewEncoder(tmpFile).Encode(cfg)
	assert.NoError(t, err)

	flagConfigPath = tmpFile.Name()

	err = flags()
	assert.NoError(t, err)

	// Env vars override config file values
	assert.Equal(t, "localhost:9999", flagServerAddress)
	assert.Equal(t, "env_dsn", flagDatabaseDSN)
	assert.Equal(t, 100, flagStoreInterval)
	assert.Equal(t, "/env/path", flagFileStoragePath)
	assert.True(t, flagRestore)
	assert.Equal(t, "env_key", flagKey)
	assert.Equal(t, "/env/crypto.key", flagCryptoKey)
	assert.Equal(t, "192.168.0.0/24", flagTrustedSubnet)
}

func TestFlags_InvalidConfigPath(t *testing.T) {
	os.Args = []string{"cmd", "-c", "/non/existent/path/config.json"}

	err := flags()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}
