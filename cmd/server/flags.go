package main

import (
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

	envAddress       = "ADDRESS"
	envDatabaseDSN   = "DATABASE_DSN"
	envStoreInterval = "STORE_INTERVAL"
	envFilePath      = "FILE_STORAGE_PATH"
	envRestore       = "RESTORE"
	envKey           = "KEY"

	defaultAddress       = ":8080"
	defaultDatabaseDSN   = ""
	defaultStoreInterval = 300
	defaultFilePath      = ""
	defaultRestore       = false
	defaultKey           = ""

	descAddress       = "address and port to run server"
	descDatabaseDSN   = "dsn for database connection"
	descStoreInterval = "interval (in seconds) to store data"
	descFilePath      = "path to store files"
	descRestore       = "whether to restore data from backup"
	descKey           = "key used for SHA256 hashing"
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
)

func parseFlags() {
	flag.StringVar(&flagServerAddress, flagAddressName, defaultAddress, descAddress)
	flag.StringVar(&flagDatabaseDSN, flagDatabaseDSNName, defaultDatabaseDSN, descDatabaseDSN)
	flag.IntVar(&flagStoreInterval, flagStoreIntervalName, defaultStoreInterval, descStoreInterval)
	flag.StringVar(&flagFileStoragePath, flagFilePathName, defaultFilePath, descFilePath)
	flag.BoolVar(&flagRestore, flagRestoreName, defaultRestore, descRestore)
	flag.StringVar(&flagKey, flagKeyName, defaultKey, descKey)

	flag.Parse()

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
}
