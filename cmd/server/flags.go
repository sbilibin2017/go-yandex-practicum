package main

import (
	"flag"
	"os"
	"strconv"
)

const (
	flagServerAddress   = "a"
	flagDatabaseDSN     = "d"
	flagStoreInterval   = "i"
	flagFileStoragePath = "f"
	flagRestore         = "r"
	flagLogLevel        = "l"

	envServerAddress   = "ADDRESS"
	envDatabaseDSN     = "DATABASE_DSN"
	envStoreInterval   = "STORE_INTERVAL"
	envFileStoragePath = "FILE_STORAGE_PATH"
	envRestore         = "RESTORE"
	envLogLevel        = "LOG_LEVEL"

	defaultServerAddress   = ":8080"
	defaultDatabaseDSN     = ""
	defaultLogLevel        = "info"
	defaultStoreInterval   = 300
	defaultFileStoragePath = ""
	defaultRestore         = false

	flagServerAddressUsage   = "address and port to run server"
	flagDatabaseDSNUse       = "DSN (Data Source Name) for database connection"
	flagStoreIntervalUsage   = "interval (in seconds) to store data"
	flagFileStoragePathUsage = "path to store files"
	flagRestoreUsage         = "whether to restore data from backup"
	flagLogLevelUsage        = "logging level (e.g., info, debug, error)"
)

type options struct {
	ServerAddress   string
	LogLevel        string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
}

var opts options

func parseFlags() *options {
	flag.StringVar(&opts.ServerAddress, flagServerAddress, defaultServerAddress, flagServerAddressUsage)
	flag.StringVar(&opts.DatabaseDSN, flagDatabaseDSN, defaultDatabaseDSN, flagDatabaseDSNUse)
	flag.IntVar(&opts.StoreInterval, flagStoreInterval, defaultStoreInterval, flagStoreIntervalUsage)
	flag.StringVar(&opts.FileStoragePath, flagFileStoragePath, defaultFileStoragePath, flagFileStoragePathUsage)
	flag.BoolVar(&opts.Restore, flagRestore, defaultRestore, flagRestoreUsage)
	flag.StringVar(&opts.LogLevel, flagLogLevel, defaultLogLevel, flagLogLevelUsage)

	flag.Parse()

	if val := os.Getenv(envServerAddress); val != "" {
		opts.ServerAddress = val
	}
	if val := os.Getenv(envDatabaseDSN); val != "" {
		opts.DatabaseDSN = val
	}
	if val := os.Getenv(envStoreInterval); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			opts.StoreInterval = v
		}
	}
	if val := os.Getenv(envFileStoragePath); val != "" {
		opts.FileStoragePath = val
	}
	if val := os.Getenv(envRestore); val != "" {
		if v, err := strconv.ParseBool(val); err == nil {
			opts.Restore = v
		}
	}
	if val := os.Getenv(envLogLevel); val != "" {
		opts.LogLevel = val
	}

	return &opts
}
