package main

import (
	"flag"
	"os"
	"strconv"
)

var (
	serverAddress   string
	databaseDSN     string
	storeInterval   int
	fileStoragePath string
	restore         bool
	key             string
	header          string
	logLevel        string
)

func parseFlags() {
	flag.StringVar(&serverAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&databaseDSN, "d", "", "DSN (Data Source Name) for database connection")
	flag.IntVar(&storeInterval, "i", 300, "interval (in seconds) to store data")
	flag.StringVar(&fileStoragePath, "f", "", "path to store files")
	flag.BoolVar(&restore, "r", false, "whether to restore data from backup")
	flag.StringVar(&key, "k", "", "key used for SHA256 hashing")

	flag.Parse()

	if env := os.Getenv("ADDRESS"); env != "" {
		serverAddress = env
	}
	if env := os.Getenv("DATABASE_DSN"); env != "" {
		databaseDSN = env
	}
	if env := os.Getenv("STORE_INTERVAL"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			storeInterval = v
		}
	}
	if env := os.Getenv("FILE_STORAGE_PATH"); env != "" {
		fileStoragePath = env
	}
	if env := os.Getenv("RESTORE"); env != "" {
		if v, err := strconv.ParseBool(env); err == nil {
			restore = v
		}
	}
	if env := os.Getenv("KEY"); env != "" {
		key = env
	}

	header = "HashSHA256"
	logLevel = "info"
}
