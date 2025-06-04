package main

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"time"
)

const (
	flagAddressName        = "a"
	flagPollIntervalName   = "p"
	flagReportIntervalName = "r"
	flagKeyName            = "k"
	flagRateLimitName      = "l"
	flagCryptoKeyName      = "crypto-key"
	flagConfigName         = "c"
	flagConfigNameLong     = "config"

	envAddress        = "ADDRESS"
	envPollInterval   = "POLL_INTERVAL"
	envReportInterval = "REPORT_INTERVAL"
	envKey            = "KEY"
	envRateLimit      = "RATE_LIMIT"
	envCryptoKey      = "CRYPTO_KEY"
	envConfig         = "CONFIG"

	defaultAddress        = "http://localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultKey            = ""
	defaultRateLimit      = 0
	defaultCryptoKey      = ""
	defaultConfig         = ""

	descAddress        = "Metrics server address"
	descPollInterval   = "Poll interval in seconds"
	descReportInterval = "Report interval in seconds"
	descKey            = "Key for HMAC SHA256 hash"
	descRateLimit      = "Max number of concurrent outgoing requests"
	descCryptoKey      = "Path to public key file for encryption"
	descConfig         = "Path to config file"
)

const (
	batchSize     = 100
	logLevel      = "info"
	hashKeyHeader = "HashSHA256"
	emptyString   = ""
)

var (
	flagServerAddress  string
	flagPollInterval   int
	flagReportInterval int
	flagKey            string
	flagRateLimit      int
	flagCryptoKey      string
	flagConfigPath     string
)

func parseFlags() error {
	flag.StringVar(&flagServerAddress, flagAddressName, defaultAddress, descAddress)
	flag.IntVar(&flagPollInterval, flagPollIntervalName, defaultPollInterval, descPollInterval)
	flag.IntVar(&flagReportInterval, flagReportIntervalName, defaultReportInterval, descReportInterval)
	flag.StringVar(&flagKey, flagKeyName, defaultKey, descKey)
	flag.IntVar(&flagRateLimit, flagRateLimitName, defaultRateLimit, descRateLimit)
	flag.StringVar(&flagCryptoKey, flagCryptoKeyName, defaultCryptoKey, descCryptoKey)
	flag.StringVar(&flagConfigPath, flagConfigName, defaultConfig, descConfig)
	flag.StringVar(&flagConfigPath, flagConfigNameLong, defaultConfig, descConfig)

	flag.Parse()

	err := loadConfigFile(flagConfigPath)

	if err != nil {
		return err
	}

	if env := os.Getenv(envAddress); env != emptyString {
		flagServerAddress = env
	}
	if env := os.Getenv(envPollInterval); env != emptyString {
		if v, err := strconv.Atoi(env); err == nil {
			flagPollInterval = v
		}
	}
	if env := os.Getenv(envReportInterval); env != emptyString {
		if v, err := strconv.Atoi(env); err == nil {
			flagReportInterval = v
		}
	}
	if env := os.Getenv(envKey); env != emptyString {
		flagKey = env
	}
	if env := os.Getenv(envRateLimit); env != emptyString {
		if v, err := strconv.Atoi(env); err == nil {
			flagRateLimit = v
		}
	}
	if env := os.Getenv(envCryptoKey); env != "" {
		flagCryptoKey = env
	}

	return nil
}

func loadConfigFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var cfg struct {
		Address        string `json:"address"`
		ReportInterval string `json:"report_interval"`
		PollInterval   string `json:"poll_interval"`
		CryptoKey      string `json:"crypto_key"`
	}

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return err
	}

	if cfg.Address != "" {
		flagServerAddress = cfg.Address
	}
	if cfg.ReportInterval != "" {
		if d, err := time.ParseDuration(cfg.ReportInterval); err == nil {
			flagReportInterval = int(d.Seconds())
		}
	}
	if cfg.PollInterval != "" {
		if d, err := time.ParseDuration(cfg.PollInterval); err == nil {
			flagPollInterval = int(d.Seconds())
		}
	}
	if cfg.CryptoKey != "" {
		flagCryptoKey = cfg.CryptoKey
	}

	return nil
}
