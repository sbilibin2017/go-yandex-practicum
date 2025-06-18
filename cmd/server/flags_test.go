package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetTestFlags() {
	flagServerAddress = ":8080"
	flagDatabaseDSN = ""
	flagStoreInterval = 300
	flagFileStoragePath = ""
	flagRestore = false
	flagKey = ""
	flagCryptoKey = ""
	flagConfigPath = ""
}

func clearEnv(keys ...string) {
	for _, k := range keys {
		os.Unsetenv(k)
	}
}

func resetFlags() {
	flagServerAddress = ":8080"
	flagDatabaseDSN = ""
	flagStoreInterval = 300
	flagFileStoragePath = ""
	flagRestore = false
	flagKey = ""
	flagCryptoKey = ""
	flagConfigPath = ""
}

func TestParseEnv(t *testing.T) {
	tests := []struct {
		name          string
		env           map[string]string
		wantAddress   string
		wantDSN       string
		wantInterval  int
		wantFilePath  string
		wantRestore   bool
		wantKey       string
		wantCryptoKey string
		wantConfig    string
	}{
		{
			name:          "empty env",
			env:           nil,
			wantAddress:   ":8080",
			wantDSN:       "",
			wantInterval:  300,
			wantFilePath:  "",
			wantRestore:   false,
			wantKey:       "",
			wantCryptoKey: "",
			wantConfig:    "",
		},
		{
			name: "all env set with valid values",
			env: map[string]string{
				"ADDRESS":           "127.0.0.1:9000",
				"DATABASE_DSN":      "user:pass@/dbname",
				"STORE_INTERVAL":    "120",
				"FILE_STORAGE_PATH": "/tmp/files",
				"RESTORE":           "true",
				"KEY":               "envkey",
				"CRYPTO_KEY":        "/env/crypto.pem",
				"CONFIG":            "/env/config.json",
			},
			wantAddress:   "127.0.0.1:9000",
			wantDSN:       "user:pass@/dbname",
			wantInterval:  120,
			wantFilePath:  "/tmp/files",
			wantRestore:   true,
			wantKey:       "envkey",
			wantCryptoKey: "/env/crypto.pem",
			wantConfig:    "/env/config.json",
		},
		{
			name: "invalid int and bool env values",
			env: map[string]string{
				"STORE_INTERVAL": "invalid",
				"RESTORE":        "notabool",
			},
			wantAddress:   ":8080",
			wantDSN:       "",
			wantInterval:  300, // default because invalid input
			wantFilePath:  "",
			wantRestore:   false, // default because invalid input
			wantKey:       "",
			wantCryptoKey: "",
			wantConfig:    "",
		},
		{
			name: "partial env set",
			env: map[string]string{
				"ADDRESS":        "192.168.1.1",
				"STORE_INTERVAL": "10",
				"RESTORE":        "true",
			},
			wantAddress:   "192.168.1.1",
			wantDSN:       "",
			wantInterval:  10,
			wantFilePath:  "",
			wantRestore:   true,
			wantKey:       "",
			wantCryptoKey: "",
			wantConfig:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()

			// Clear all relevant env vars before test
			clearEnv("ADDRESS", "DATABASE_DSN", "STORE_INTERVAL", "FILE_STORAGE_PATH", "RESTORE", "KEY", "CRYPTO_KEY", "CONFIG")

			// Set environment variables for this test case
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			// Call the function under test
			parseEnv()

			assert.Equal(t, tt.wantAddress, flagServerAddress)
			assert.Equal(t, tt.wantDSN, flagDatabaseDSN)
			assert.Equal(t, tt.wantInterval, flagStoreInterval)
			assert.Equal(t, tt.wantFilePath, flagFileStoragePath)
			assert.Equal(t, tt.wantRestore, flagRestore)
			assert.Equal(t, tt.wantKey, flagKey)
			assert.Equal(t, tt.wantCryptoKey, flagCryptoKey)
			assert.Equal(t, tt.wantConfig, flagConfigPath)

			// Cleanup env vars
			clearEnv("ADDRESS", "DATABASE_DSN", "STORE_INTERVAL", "FILE_STORAGE_PATH", "RESTORE", "KEY", "CRYPTO_KEY", "CONFIG")
		})
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "config-*.json") // use os.CreateTemp
	assert.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)

	err = tmpFile.Close()
	assert.NoError(t, err)

	return tmpFile.Name()
}

func TestParseConfigFile(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		initialFlags  map[string]interface{}
		wantFlags     map[string]interface{}
		wantErr       bool
	}{
		{
			name:          "no config path returns nil",
			configContent: "",
			initialFlags:  nil,
			wantFlags: map[string]interface{}{
				"flagServerAddress":   ":8080",
				"flagDatabaseDSN":     "",
				"flagStoreInterval":   300,
				"flagFileStoragePath": "",
				"flagRestore":         false,
				"flagKey":             "",
				"flagCryptoKey":       "",
			},
			wantErr: false,
		},
		{
			name: "valid config overrides defaults",
			configContent: `{
				"address": "127.0.0.1:9090",
				"database_dsn": "user:pass@/dbname",
				"store_interval": 60,
				"store_file": "/tmp/store.json",
				"restore": true,
				"key": "secretkey",
				"crypto_key": "/tmp/crypto.pem"
			}`,
			initialFlags: nil,
			wantFlags: map[string]interface{}{
				"flagServerAddress":   "127.0.0.1:9090",
				"flagDatabaseDSN":     "user:pass@/dbname",
				"flagStoreInterval":   60,
				"flagFileStoragePath": "/tmp/store.json",
				"flagRestore":         true,
				"flagKey":             "secretkey",
				"flagCryptoKey":       "/tmp/crypto.pem",
			},
			wantErr: false,
		},
		{
			name:          "invalid JSON returns error",
			configContent: `{ invalid json `,
			initialFlags:  nil,
			wantErr:       true,
		},
		{
			name: "no override if flags already set",
			configContent: `{
				"address": "127.0.0.1:9090",
				"database_dsn": "user:pass@/dbname",
				"store_interval": 60,
				"store_file": "/tmp/store.json",
				"restore": true,
				"key": "secretkey",
				"crypto_key": "/tmp/crypto.pem"
			}`,
			initialFlags: map[string]interface{}{
				"flagServerAddress":   "0.0.0.0:8080",
				"flagDatabaseDSN":     "existingDSN",
				"flagStoreInterval":   120,
				"flagFileStoragePath": "/existing/path",
				"flagRestore":         true,
				"flagKey":             "existingkey",
				"flagCryptoKey":       "/existing/crypto.pem",
			},
			wantFlags: map[string]interface{}{
				"flagServerAddress":   "0.0.0.0:8080",
				"flagDatabaseDSN":     "existingDSN",
				"flagStoreInterval":   120,
				"flagFileStoragePath": "/existing/path",
				"flagRestore":         true,
				"flagKey":             "existingkey",
				"flagCryptoKey":       "/existing/crypto.pem",
			},
			wantErr: false,
		},
		{
			name: "partial override",
			configContent: `{
				"address": "192.168.1.1:7070",
				"key": "partialkey"
			}`,
			initialFlags: map[string]interface{}{
				"flagStoreInterval": 120,
			},
			wantFlags: map[string]interface{}{
				"flagServerAddress":   "192.168.1.1:7070",
				"flagDatabaseDSN":     "",
				"flagStoreInterval":   120,
				"flagFileStoragePath": "",
				"flagRestore":         false,
				"flagKey":             "partialkey",
				"flagCryptoKey":       "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetTestFlags()

			// Set initial flags if any
			if tt.initialFlags != nil {
				if v, ok := tt.initialFlags["flagServerAddress"]; ok {
					flagServerAddress = v.(string)
				}
				if v, ok := tt.initialFlags["flagDatabaseDSN"]; ok {
					flagDatabaseDSN = v.(string)
				}
				if v, ok := tt.initialFlags["flagStoreInterval"]; ok {
					flagStoreInterval = v.(int)
				}
				if v, ok := tt.initialFlags["flagFileStoragePath"]; ok {
					flagFileStoragePath = v.(string)
				}
				if v, ok := tt.initialFlags["flagRestore"]; ok {
					flagRestore = v.(bool)
				}
				if v, ok := tt.initialFlags["flagKey"]; ok {
					flagKey = v.(string)
				}
				if v, ok := tt.initialFlags["flagCryptoKey"]; ok {
					flagCryptoKey = v.(string)
				}
			}

			if tt.configContent != "" {
				path := writeTempConfig(t, tt.configContent)
				defer os.Remove(path)
				flagConfigPath = path
			} else {
				flagConfigPath = ""
			}

			err := parseConfigFile()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Check flags
			assert.Equal(t, tt.wantFlags["flagServerAddress"], flagServerAddress)
			assert.Equal(t, tt.wantFlags["flagDatabaseDSN"], flagDatabaseDSN)
			assert.Equal(t, tt.wantFlags["flagStoreInterval"], flagStoreInterval)
			assert.Equal(t, tt.wantFlags["flagFileStoragePath"], flagFileStoragePath)
			assert.Equal(t, tt.wantFlags["flagRestore"], flagRestore)
			assert.Equal(t, tt.wantFlags["flagKey"], flagKey)
			assert.Equal(t, tt.wantFlags["flagCryptoKey"], flagCryptoKey)
		})
	}
}
