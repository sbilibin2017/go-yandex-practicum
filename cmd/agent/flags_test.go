package main

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func resetFlags() {
	// Reset CommandLine to new FlagSet with ContinueOnError (to match parseFlags)
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
}

func clearEnv(keys ...string) {
	for _, k := range keys {
		os.Unsetenv(k)
	}
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantAddress   string
		wantPoll      int
		wantReport    int
		wantKey       string
		wantRateLimit int
		wantCryptoKey string
		wantConfig    string
	}{
		{
			name:          "default values",
			args:          []string{"cmd"},
			wantAddress:   "http://localhost:8080",
			wantPoll:      2,
			wantReport:    10,
			wantKey:       "",
			wantRateLimit: 0,
			wantCryptoKey: "",
			wantConfig:    "",
		},
		{
			name: "all flags set",
			args: []string{
				"cmd",
				"-a", "http://example.com",
				"-p", "5",
				"-r", "15",
				"-k", "mykey",
				"-l", "10",
				"--crypto-key", "/tmp/key.pem",
				"-c", "/tmp/config.json",
			},
			wantAddress:   "http://example.com",
			wantPoll:      5,
			wantReport:    15,
			wantKey:       "mykey",
			wantRateLimit: 10,
			wantCryptoKey: "/tmp/key.pem",
			wantConfig:    "/tmp/config.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			os.Args = tt.args

			// Reset vars before parsing
			flagServerAddress = ""
			flagPollInterval = 0
			flagReportInterval = 0
			flagKey = ""
			flagRateLimit = 0
			flagCryptoKey = ""
			flagConfigPath = ""

			err := parseFlags()
			assert.NoError(t, err)

			assert.Equal(t, tt.wantAddress, flagServerAddress)
			assert.Equal(t, tt.wantPoll, flagPollInterval)
			assert.Equal(t, tt.wantReport, flagReportInterval)
			assert.Equal(t, tt.wantKey, flagKey)
			assert.Equal(t, tt.wantRateLimit, flagRateLimit)
			assert.Equal(t, tt.wantCryptoKey, flagCryptoKey)
			assert.Equal(t, tt.wantConfig, flagConfigPath)
		})
	}
}

func TestParseEnv(t *testing.T) {
	tests := []struct {
		name          string
		env           map[string]string
		wantAddress   string
		wantPoll      int
		wantReport    int
		wantKey       string
		wantRateLimit int
		wantCryptoKey string
		wantConfig    string
	}{
		{
			name:        "empty env",
			env:         nil,
			wantAddress: "",
			wantPoll:    0,
			wantReport:  0,
			wantKey:     "",
		},
		{
			name: "all env set",
			env: map[string]string{
				"ADDRESS":         "http://env.com",
				"POLL_INTERVAL":   "7",
				"REPORT_INTERVAL": "20",
				"KEY":             "envkey",
				"RATE_LIMIT":      "5",
				"CRYPTO_KEY":      "/env/key.pem",
				"CONFIG":          "/env/config.json",
			},
			wantAddress:   "http://env.com",
			wantPoll:      7,
			wantReport:    20,
			wantKey:       "envkey",
			wantRateLimit: 5,
			wantCryptoKey: "/env/key.pem",
			wantConfig:    "/env/config.json",
		},
		{
			name: "invalid int env values",
			env: map[string]string{
				"POLL_INTERVAL":   "bad",
				"REPORT_INTERVAL": "bad",
				"RATE_LIMIT":      "bad",
			},
			wantPoll:      0,
			wantReport:    0,
			wantRateLimit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv("ADDRESS", "POLL_INTERVAL", "REPORT_INTERVAL", "KEY", "RATE_LIMIT", "CRYPTO_KEY", "CONFIG")

			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			// Reset vars before parsing
			flagServerAddress = ""
			flagPollInterval = 0
			flagReportInterval = 0
			flagKey = ""
			flagRateLimit = 0
			flagCryptoKey = ""
			flagConfigPath = ""

			err := parseEnvs()
			assert.NoError(t, err)

			assert.Equal(t, tt.wantAddress, flagServerAddress)
			assert.Equal(t, tt.wantPoll, flagPollInterval)
			assert.Equal(t, tt.wantReport, flagReportInterval)
			assert.Equal(t, tt.wantKey, flagKey)
			assert.Equal(t, tt.wantRateLimit, flagRateLimit)
			assert.Equal(t, tt.wantCryptoKey, flagCryptoKey)
			assert.Equal(t, tt.wantConfig, flagConfigPath)

			clearEnv("ADDRESS", "POLL_INTERVAL", "REPORT_INTERVAL", "KEY", "RATE_LIMIT", "CRYPTO_KEY", "CONFIG")
		})
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "config-*.json")
	assert.NoError(t, err)
	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)
	return tmpFile.Name()
}

func resetTestFlags() {
	flagServerAddress = "http://localhost:8080"
	flagReportInterval = 10
	flagPollInterval = 2
	flagKey = ""
	flagRateLimit = 0
	flagCryptoKey = ""
	flagConfigPath = ""
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
				"flagServerAddress":  "http://localhost:8080",
				"flagReportInterval": 10,
				"flagPollInterval":   2,
				"flagKey":            "",
				"flagRateLimit":      0,
				"flagCryptoKey":      "",
			},
			wantErr: false,
		},
		{
			name: "valid config overrides defaults",
			configContent: `{
				"server_address": "http://example.com",
				"report_interval": 15,
				"poll_interval": 5,
				"key": "secret",
				"rate_limit": 7,
				"crypto_key_path": "/path/to/crypto"
			}`,
			initialFlags: nil,
			wantFlags: map[string]interface{}{
				"flagServerAddress":  "http://example.com",
				"flagReportInterval": 15,
				"flagPollInterval":   5,
				"flagKey":            "secret",
				"flagRateLimit":      7,
				"flagCryptoKey":      "/path/to/crypto",
			},
			wantErr: false,
		},
		{
			name:          "invalid JSON returns error",
			configContent: `{invalid json}`,
			initialFlags:  nil,
			wantErr:       true,
		},
		{
			name: "no override if flags already set",
			configContent: `{
				"server_address": "http://example.com",
				"report_interval": 15,
				"poll_interval": 5,
				"key": "secret",
				"rate_limit": 7,
				"crypto_key_path": "/path/to/crypto"
			}`,
			initialFlags: map[string]interface{}{
				"flagServerAddress":  "http://myserver",
				"flagReportInterval": 20,
				"flagPollInterval":   10,
				"flagKey":            "mykey",
				"flagRateLimit":      9,
				"flagCryptoKey":      "/my/crypto",
			},
			wantFlags: map[string]interface{}{
				"flagServerAddress":  "http://myserver",
				"flagReportInterval": 20,
				"flagPollInterval":   10,
				"flagKey":            "mykey",
				"flagRateLimit":      9,
				"flagCryptoKey":      "/my/crypto",
			},
			wantErr: false,
		},
		{
			name: "partial override",
			configContent: `{
				"server_address": "http://partial.com",
				"key": "partialkey"
			}`,
			initialFlags: map[string]interface{}{
				"flagReportInterval": 20,
			},
			wantFlags: map[string]interface{}{
				"flagServerAddress":  "http://partial.com",
				"flagReportInterval": 20,
				"flagPollInterval":   2,
				"flagKey":            "partialkey",
				"flagRateLimit":      0,
				"flagCryptoKey":      "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetTestFlags()

			if tt.initialFlags != nil {
				if v, ok := tt.initialFlags["flagServerAddress"]; ok {
					flagServerAddress = v.(string)
				}
				if v, ok := tt.initialFlags["flagReportInterval"]; ok {
					flagReportInterval = v.(int)
				}
				if v, ok := tt.initialFlags["flagPollInterval"]; ok {
					flagPollInterval = v.(int)
				}
				if v, ok := tt.initialFlags["flagKey"]; ok {
					flagKey = v.(string)
				}
				if v, ok := tt.initialFlags["flagRateLimit"]; ok {
					flagRateLimit = v.(int)
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

			for k, want := range tt.wantFlags {
				switch k {
				case "flagServerAddress":
					assert.Equal(t, want, flagServerAddress)
				case "flagReportInterval":
					assert.Equal(t, want, flagReportInterval)
				case "flagPollInterval":
					assert.Equal(t, want, flagPollInterval)
				case "flagKey":
					assert.Equal(t, want, flagKey)
				case "flagRateLimit":
					assert.Equal(t, want, flagRateLimit)
				case "flagCryptoKey":
					assert.Equal(t, want, flagCryptoKey)
				}
			}
		})
	}
}
