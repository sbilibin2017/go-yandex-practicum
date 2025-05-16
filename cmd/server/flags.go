package main

import (
	"flag"
	"os"
	"strconv"
)

var (
	flagServerAddress   string
	flagDatabaseDSN     string
	flagStoreInterval   int
	flagFileStoragePath string
	flagRestore         bool
	flagKey             string
	flagLogLevel        string
)

func parseFlags() {
	flag.StringVar(&flagServerAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&flagDatabaseDSN, "d", "", "DSN (Data Source Name) for database connection")
	flag.IntVar(&flagStoreInterval, "i", 300, "interval (in seconds) to store data")
	flag.StringVar(&flagFileStoragePath, "f", "", "path to store files")
	flag.BoolVar(&flagRestore, "r", false, "whether to restore data from backup")
	flag.StringVar(&flagKey, "k", "", "key used for SHA256 hashing")
	flag.StringVar(&flagLogLevel, "l", "info", "logging level (e.g., info, debug, error)")

	flag.Parse()

	if val := os.Getenv("ADDRESS"); val != "" {
		flagServerAddress = val
	}
	if val := os.Getenv("DATABASE_DSN"); val != "" {
		flagDatabaseDSN = val
	}
	if val := os.Getenv("STORE_INTERVAL"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			flagStoreInterval = v
		}
	}
	if val := os.Getenv("FILE_STORAGE_PATH"); val != "" {
		flagFileStoragePath = val
	}
	if val := os.Getenv("RESTORE"); val != "" {
		if v, err := strconv.ParseBool(val); err == nil {
			flagRestore = v
		}
	}
	if val := os.Getenv("KEY"); val != "" {
		flagKey = val
	}
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		flagLogLevel = val
	}
}
