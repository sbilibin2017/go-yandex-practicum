package main

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
)

const (
	flagAddressName       = "a"
	flagDatabaseDSNName   = "d"
	flagStoreIntervalName = "i"
	flagFilePathName      = "f"
	flagRestoreName       = "r"
	flagKeyName           = "k"
	flagCryptoKeyName     = "crypto-key"
	flagConfigName        = "c"
	flagConfigNameLong    = "config"

	envAddress       = "ADDRESS"
	envDatabaseDSN   = "DATABASE_DSN"
	envStoreInterval = "STORE_INTERVAL"
	envFilePath      = "FILE_STORAGE_PATH"
	envRestore       = "RESTORE"
	envKey           = "KEY"
	envCryptoKey     = "CRYPTO_KEY"
	envConfig        = "CONFIG"

	defaultAddress       = ":8080"
	defaultDatabaseDSN   = ""
	defaultStoreInterval = 300
	defaultFilePath      = ""
	defaultRestore       = false
	defaultKey           = ""
	defaultCryptoKey     = ""
	defaultConfig        = ""

	descAddress       = "address and port to run server"
	descDatabaseDSN   = "dsn for database connection"
	descStoreInterval = "interval (in seconds) to store data"
	descFilePath      = "path to store files"
	descRestore       = "whether to restore data from backup"
	descKey           = "key used for SHA256 hashing"
	descCryptoKey     = "path to private key file for encryption"
	descConfig        = "path to config file"
)

const (
	hashKeyHeader = "HashSHA256"
	logLevel      = "info"
	emptyString   = ""
)

var (
	flagServerAddress   string
	flagDatabaseDSN     string
	flagStoreInterval   int
	flagFileStoragePath string
	flagRestore         bool
	flagKey             string
	flagCryptoKey       string
	flagConfigPath      string
)

func parseFlags() error {
	flag.StringVar(&flagServerAddress, flagAddressName, defaultAddress, descAddress)
	flag.StringVar(&flagDatabaseDSN, flagDatabaseDSNName, defaultDatabaseDSN, descDatabaseDSN)
	flag.IntVar(&flagStoreInterval, flagStoreIntervalName, defaultStoreInterval, descStoreInterval)
	flag.StringVar(&flagFileStoragePath, flagFilePathName, defaultFilePath, descFilePath)
	flag.BoolVar(&flagRestore, flagRestoreName, defaultRestore, descRestore)
	flag.StringVar(&flagKey, flagKeyName, defaultKey, descKey)
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
	if env := os.Getenv(envDatabaseDSN); env != emptyString {
		flagDatabaseDSN = env
	}
	if env := os.Getenv(envStoreInterval); env != emptyString {
		if v, err := strconv.Atoi(env); err == nil {
			flagStoreInterval = v
		}
	}
	if env := os.Getenv(envFilePath); env != emptyString {
		flagFileStoragePath = env
	}
	if env := os.Getenv(envRestore); env != emptyString {
		if v, err := strconv.ParseBool(env); err == nil {
			flagRestore = v
		}
	}
	if env := os.Getenv(envKey); env != emptyString {
		flagKey = env
	}
	if env := os.Getenv(envCryptoKey); env != "" {
		flagCryptoKey = env
	}

	return nil
}

func loadConfigFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var cfg struct {
		Address       string `json:"address"`
		DatabaseDSN   string `json:"database_dsn"`
		StoreInterval int    `json:"store_interval"`
		FilePath      string `json:"store_file"`
		Restore       bool   `json:"restore"`
		Key           string `json:"key"`
		CryptoKey     string `json:"crypto_key"`
	}

	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return err
	}

	if cfg.Address != "" && flagServerAddress == defaultAddress {
		flagServerAddress = cfg.Address
	}
	if cfg.DatabaseDSN != "" && flagDatabaseDSN == defaultDatabaseDSN {
		flagDatabaseDSN = cfg.DatabaseDSN
	}
	if cfg.StoreInterval != 0 && flagStoreInterval == defaultStoreInterval {
		flagStoreInterval = cfg.StoreInterval
	}
	if cfg.FilePath != "" && flagFileStoragePath == defaultFilePath {
		flagFileStoragePath = cfg.FilePath
	}
	if !flagRestore {
		flagRestore = cfg.Restore
	}
	if cfg.Key != "" && flagKey == defaultKey {
		flagKey = cfg.Key
	}
	if cfg.CryptoKey != "" && flagCryptoKey == defaultCryptoKey {
		flagCryptoKey = cfg.CryptoKey
	}

	return nil
}
